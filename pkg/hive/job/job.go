package job

import (
	"context"
	"errors"
	"runtime/pprof"
	"strconv"
	"sync"
	"time"

	"github.com/ccfish2/infra/pkg/hive"
	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/hive/internal"
	"github.com/ccfish2/infra/pkg/inctimer"
	"github.com/ccfish2/infra/pkg/lock"
	"github.com/ccfish2/infra/pkg/spanstat"
	"github.com/ccfish2/infra/pkg/stream"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/util/workqueue"
)

var Cell = cell.Module(
	"jobs",
	"Jobs",
	cell.Provide(newRegistry),
	cell.Metric(newJobMetrics),
)

type Registry interface {
	NewGroup(scope cell.Scope, opts ...groupOpt) Group
}

type registry struct {
	logger     logrus.FieldLogger
	shutdowner hive.Shutdowner

	metrics *jobMetrics

	mu     lock.Mutex
	groups []Group
}

func newRegistry(
	logger logrus.FieldLogger,
	shutdowner hive.Shutdowner,
	metrics *jobMetrics,
) Registry {
	return &registry{
		logger:     logger,
		shutdowner: shutdowner,
		metrics:    metrics,
	}
}

func (c *registry) NewGroup(scope cell.Scope, opts ...groupOpt) Group {
	c.mu.Lock()
	defer c.mu.Unlock()

	var options options
	options.logger = c.logger
	options.shutdowner = c.shutdowner
	options.metrics = c.metrics

	for _, opt := range opts {
		opt(&options)
	}

	g := &group{
		options: options,
		wg:      &sync.WaitGroup{},
		scope:   scope,
	}

	c.groups = append(c.groups, g)

	return g
}

type Group interface {
	Add(...Job)
	Scoped(name string) ScopedGroup
	cell.HookInterface
}

type Job interface {
	start(ctx context.Context, wg *sync.WaitGroup, scope cell.Scope, options options)
}

type group struct {
	options options

	wg *sync.WaitGroup

	mu         lock.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	queuedJobs []Job

	scope cell.Scope
}

type options struct {
	pprofLabels pprof.LabelSet
	logger      logrus.FieldLogger
	shutdowner  hive.Shutdowner
	metrics     *jobMetrics
}

type groupOpt func(o *options)

func WithLogger(logger logrus.FieldLogger) groupOpt {
	return func(o *options) {
		o.logger = logger
	}
}

func WithPprofLabels(pprofLabels pprof.LabelSet) groupOpt {
	return func(o *options) {
		o.pprofLabels = pprofLabels
	}
}

var _ cell.HookInterface = (*group)(nil)

// Start implements the cell.HookInterface interface
func (jg *group) Start(_ cell.HookContext) error {
	jg.mu.Lock()
	defer jg.mu.Unlock()

	jg.ctx, jg.cancel = context.WithCancel(context.Background())

	jg.wg.Add(len(jg.queuedJobs))
	for _, job := range jg.queuedJobs {
		pprof.Do(jg.ctx, jg.options.pprofLabels, func(ctx context.Context) {
			go job.start(ctx, jg.wg, jg.scope, jg.options)
		})
	}
	jg.queuedJobs = nil

	return nil
}

// Stop implements the cell.HookInterface interface
func (jg *group) Stop(stopCtx cell.HookContext) error {
	jg.mu.Lock()
	defer jg.mu.Unlock()

	done := make(chan struct{})
	go func() {
		jg.wg.Wait()
		close(done)
	}()

	jg.cancel()

	select {
	case <-stopCtx.Done():
		jg.options.logger.Error("Stop hook context expired before job group was done")
	case <-done:
	}

	return nil
}

func (jg *group) Add(jobs ...Job) {
	jg.add(jg.scope, jobs...)
}

func (jg *group) add(scope cell.Scope, jobs ...Job) {
	jg.mu.Lock()
	defer jg.mu.Unlock()

	// The context is only set once the group has been started. If we have not yet started, queue the jobs.
	if jg.ctx == nil {
		jg.queuedJobs = append(jg.queuedJobs, jobs...)
		return
	}

	for _, j := range jobs {
		jg.wg.Add(1)
		pprof.Do(jg.ctx, jg.options.pprofLabels, func(ctx context.Context) {
			go j.start(ctx, jg.wg, scope, jg.options)
		})
	}
}

func (jg *group) Scoped(name string) ScopedGroup {
	return &scopedGroup{
		group: jg,
		scope: cell.GetSubScope(jg.scope, name),
	}
}

type ScopedGroup interface {
	Add(jobs ...Job)
}

type scopedGroup struct {
	group *group
	scope cell.Scope
}

func (sg *scopedGroup) Add(jobs ...Job) {
	sg.group.add(sg.scope, jobs...)
}

func OneShot(name string, fn OneShotFunc, opts ...jobOneShotOpt) Job {
	if fn == nil {
		panic("`fn` must not be nil")
	}

	job := &jobOneShot{
		name: name,
		fn:   fn,
		opts: opts,
	}

	return job
}

type jobOneShotOpt func(*jobOneShot)

func WithRetry(times int, backoff workqueue.RateLimiter) jobOneShotOpt {
	return func(jos *jobOneShot) {
		jos.retry = times
		jos.backoff = backoff
	}
}

func WithShutdown() jobOneShotOpt {
	return func(jos *jobOneShot) {
		jos.shutdownOnError = true
	}
}

func WithMetrics() jobOneShotOpt {
	return func(jos *jobOneShot) {
		jos.metrics = true
	}
}

type OneShotFunc func(ctx context.Context, health cell.HealthReporter) error

type jobOneShot struct {
	name string
	fn   OneShotFunc
	opts []jobOneShotOpt

	health cell.HealthReporter

	// If retry > 0, retry on error x times.
	retry           int
	backoff         workqueue.RateLimiter
	shutdownOnError bool
	metrics         bool
}

func (jos *jobOneShot) start(ctx context.Context, wg *sync.WaitGroup, scope cell.Scope, options options) {
	defer wg.Done()

	for _, opt := range jos.opts {
		opt(jos)
	}

	jos.health = cell.GetHealthReporter(scope, "job-"+jos.name)
	defer jos.health.Stopped("one-shot job done")

	l := options.logger.WithFields(logrus.Fields{
		"name": jos.name,
		"func": internal.FuncNameAndLocation(jos.fn),
	})

	stat := &spanstat.SpanStat{}

	timer, cancel := inctimer.New()
	defer cancel()

	var err error
	for i := 0; i <= jos.retry; i++ {
		var timeout time.Duration
		if i != 0 {
			timeout = jos.backoff.When(jos)
			l.WithFields(logrus.Fields{
				"backoff":     timeout,
				"retry-count": i,
			}).Debug("Delaying retry attempt")
		}

		select {
		case <-ctx.Done():
			return
		case <-timer.After(timeout):
		}

		l.Debug("Starting one-shot job")

		if jos.metrics {
			stat.Start()
		}

		jos.health.OK("Running")
		err = jos.fn(ctx, jos.health)

		if jos.metrics {
			sec := stat.End(true).Seconds()
			options.metrics.OneShotRunDuration.WithLabelValues(jos.name).Observe(sec)
			stat.Reset()
		}

		if err == nil {
			return
		} else if !errors.Is(err, context.Canceled) {
			jos.health.Degraded("one-shot job errored", err)
			l.WithError(err).Error("one-shot job errored")
			options.metrics.JobErrorsTotal.WithLabelValues(jos.name).Inc()
		}
	}

	if options.shutdowner != nil && jos.shutdownOnError {
		options.shutdowner.Shutdown(hive.ShutdownWithError(err))
	}
}

func Timer(name string, fn TimerFunc, interval time.Duration, opts ...timerOpt) Job {
	if fn == nil {
		panic("`fn` must not be nil")
	}

	job := &jobTimer{
		name:     name,
		fn:       fn,
		interval: interval,
		opts:     opts,
	}

	return job
}

// TimerFunc is the func type invoked by a timer job. A TimerFunc is expected to return as soon as the ctx expires.
type TimerFunc func(ctx context.Context) error

type timerOpt func(*jobTimer)

// Trigger which can be used to trigger a timer job, trigger events are coalesced.
type Trigger interface {
	_trigger()
	Trigger()
}

// NewTrigger creates a new trigger, which can be used to trigger a timer job.
func NewTrigger() *trigger {
	return &trigger{
		c: make(chan struct{}, 1),
	}
}

type trigger struct {
	c chan struct{}
}

func (t *trigger) _trigger() {}

func (t *trigger) Trigger() {
	select {
	case t.c <- struct{}{}:
	default:
	}
}

func WithTrigger(trig Trigger) timerOpt {
	return func(jt *jobTimer) {
		jt.trigger = trig.(*trigger)
	}
}

type jobTimer struct {
	name string
	fn   TimerFunc
	opts []timerOpt

	health cell.HealthReporter

	interval time.Duration
	trigger  *trigger

	// If not nil, call the shutdowner on error
	shutdown hive.Shutdowner
}

func (jt *jobTimer) start(ctx context.Context, wg *sync.WaitGroup, scope cell.Scope, options options) {
	defer wg.Done()

	for _, opt := range jt.opts {
		opt(jt)
	}

	jt.health = cell.GetHealthReporter(scope, "timer-job-"+jt.name)

	l := options.logger.WithFields(logrus.Fields{
		"name": jt.name,
		"func": internal.FuncNameAndLocation(jt.fn),
	})

	timer := time.NewTicker(jt.interval)
	defer timer.Stop()

	var triggerChan chan struct{}
	if jt.trigger != nil {
		triggerChan = jt.trigger.c
	}

	l.Debug("Starting timer job")
	jt.health.OK("Primed")

	stat := &spanstat.SpanStat{}

	for {
		select {
		case <-ctx.Done():
			jt.health.Stopped("timer job context done")
			return
		case <-timer.C:
		case <-triggerChan:
		}

		l.Debug("Timer job triggered")

		stat.Start()

		err := jt.fn(ctx)

		total := stat.End(true).Total()
		options.metrics.TimerRunDuration.WithLabelValues(jt.name).Observe(total.Seconds())
		stat.Reset()

		if err == nil {
			jt.health.OK("OK (" + total.String() + ")")
			l.Debug("Timer job finished")
		} else if !errors.Is(err, context.Canceled) {
			jt.health.Degraded("timer job errored", err)
			l.WithError(err).Error("Timer job errored")

			options.metrics.JobErrorsTotal.WithLabelValues(jt.name).Inc()
			if jt.shutdown != nil {
				jt.shutdown.Shutdown(hive.ShutdownWithError(err))
			}
		}

		if ctx.Err() != nil {
			return
		}
	}
}

func Observer[T any](name string, fn ObserverFunc[T], observable stream.Observable[T], opts ...observerOpt[T]) Job {
	if fn == nil {
		panic("`fn` must not be nil")
	}

	job := &jobObserver[T]{
		name:       name,
		fn:         fn,
		observable: observable,
		opts:       opts,
	}

	return job
}

type ObserverFunc[T any] func(ctx context.Context, event T) error

type observerOpt[T any] func(*jobObserver[T])

type jobObserver[T any] struct {
	name string
	fn   ObserverFunc[T]
	opts []observerOpt[T]

	health cell.HealthReporter

	observable stream.Observable[T]

	// If not nil, call the shutdowner on error
	shutdown hive.Shutdowner
}

func (jo *jobObserver[T]) start(ctx context.Context, wg *sync.WaitGroup, scope cell.Scope, options options) {
	defer wg.Done()

	for _, opt := range jo.opts {
		opt(jo)
	}

	jo.health = cell.GetHealthReporter(scope, "observer-job-"+jo.name)
	reportTicker := time.NewTicker(10 * time.Second)
	defer reportTicker.Stop()

	l := options.logger.WithFields(logrus.Fields{
		"name": jo.name,
		"func": internal.FuncNameAndLocation(jo.fn),
	})

	l.Debug("Observer job started")
	jo.health.OK("Primed")
	var msgCount uint64

	done := make(chan struct{})

	var (
		stat = &spanstat.SpanStat{}
		err  error
	)
	jo.observable.Observe(ctx, func(t T) {
		stat.Start()

		err := jo.fn(ctx, t)

		total := stat.End(true).Total()
		options.metrics.ObserverRunDuration.WithLabelValues(jo.name).Observe(total.Seconds())
		stat.Reset()

		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			jo.health.Degraded("observer job errored", err)
			l.WithError(err).Error("Observer job errored")
			options.metrics.JobErrorsTotal.WithLabelValues(jo.name).Inc()
			if jo.shutdown != nil {
				jo.shutdown.Shutdown(hive.ShutdownWithError(
					err,
				))
			}
			return
		}

		msgCount++

		// Don't report health for every event, only when we have not done so for a bit
		select {
		case <-reportTicker.C:
			jo.health.OK("OK (" + total.String() + ") [" + strconv.FormatUint(msgCount, 10) + "]")
		default:
		}
	}, func(e error) {
		err = e
		close(done)
	})

	<-done

	jo.health.Stopped("observer job done")
	if err != nil {
		l.WithError(err).Error("Observer job stopped with an error")
	} else {
		l.WithError(err).Debug("Observer job stopped")
	}
}
