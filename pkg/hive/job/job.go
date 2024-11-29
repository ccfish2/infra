package job

import (
	"context"
	"errors"
	"fmt"
	"runtime/pprof"
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

// NewGroup implements Registry.
func (r *registry) NewGroup(scope cell.Scope, opts ...groupOpt) Group {
	r.mu.Lock()
	defer r.mu.Unlock()

	option := options{
		logger:     r.logger,
		shutdowner: r.shutdowner,
		metrics:    r.metrics,
	}

	for _, opt := range opts {
		opt(&option)
	}

	g := &group{
		options: option,
		wg:      &sync.WaitGroup{},
		scope:   scope,
	}

	r.groups = append(r.groups, g)
	return g
}

func newRegistry(l logrus.FieldLogger, sh hive.Shutdowner, mj *jobMetrics) Registry {
	return &registry{
		logger:     l,
		shutdowner: sh,
		metrics:    mj,
	}
}

type Group interface {
	Add(...Job)
	Scoped(name string) ScopedGroup
	cell.HookInterface
}

type Job interface {
	Start(ctx context.Context, wg *sync.WaitGroup, scope cell.Scope, opt options)
}

type ScopedGroup interface {
	Add(...Job)
}

type scopedgroup struct {
	group *group
	scope cell.Scope
}

func (sg *scopedgroup) Add(jobs ...Job) {
	sg.group.add(sg.scope, jobs...)
}

type group struct {
	options    options
	wg         *sync.WaitGroup
	mu         lock.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	queuedJobs []Job

	scope cell.Scope
}

// Add implements Group.
func (g *group) Add(jobs ...Job) {
	g.add(g.scope, jobs...)
}

func (g *group) add(scope cell.Scope, jobs ...Job) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.ctx == nil {
		g.queuedJobs = append(g.queuedJobs, jobs...)
		return
	}

	for _, job := range jobs {
		pprof.Do(g.ctx, g.options.pprofLabels, func(ctx context.Context) {
			g.wg.Add(1)
			go job.Start(ctx, g.wg, scope, g.options)
		})
	}
}

// Scoped implements Group.
func (g *group) Scoped(name string) ScopedGroup {
	return &scopedgroup{
		group: g,
		scope: cell.GetSubScope(g.scope, name),
	}
}

// Start implements cell.HookInterface.
func (g *group) Start(_ cell.HookContext) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, job := range g.queuedJobs {
		g.wg.Add(1)
		pprof.Do(g.ctx, g.options.pprofLabels, func(ctx context.Context) {
			go job.Start(ctx, g.wg, g.scope, g.options)
		})
	}
	g.queuedJobs = nil
	return nil
}

// Stop implements cell.HookInterface.
func (g *group) Stop(stopedCtx cell.HookContext) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	done := make(chan interface{})
	go func() {
		g.wg.Wait()
		close(done)
	}()
	g.cancel()

	select {
	case <-stopedCtx.Done():
		return fmt.Errorf("got cancelled")
	case <-done:
	}
	return nil
}

type options struct {
	pprofLabels pprof.LabelSet
	logger      logrus.FieldLogger
	shutdowner  hive.Shutdowner
	metrics     *jobMetrics
}

type groupOpt func(o *options)

var _ cell.HookInterface = (*group)(nil)

type OneShotFunc func(ctx context.Context, heal cell.HealthReporter) error
type jobOneShotOpt func(*jobOneShot)

type jobOneShot struct {
	name   string
	fn     OneShotFunc
	opts   []jobOneShotOpt
	health cell.HealthReporter

	retry            int
	backoff          workqueue.RateLimiter
	shutdownnOnError bool
	metrics          bool
}

func OneShot(name string, fn OneShotFunc, opts ...jobOneShotOpt) Job {
	if fn == nil {
		return nil
	}
	return &jobOneShot{
		name: name,
		fn:   fn,
		opts: opts,
	}
}

func (jos *jobOneShot) Start(ctx context.Context, wg *sync.WaitGroup, scope cell.Scope, options options) {
	defer wg.Done()

	for _, opt := range jos.opts {
		opt(jos)
	}

	jos.health = cell.GetHealthReporter(scope, "job-"+jos.name)
	defer jos.health.Stopped("one-shot job stopped")

	l := options.logger.WithFields(logrus.Fields{
		"name":    jos.name,
		"funtion": internal.FuncNameAndLocation(jos.fn),
	})

	spanstat := spanstat.SpanStat{}

	timer, cancel := inctimer.New()
	defer cancel()

	var err error
	for i := 0; i < jos.retry; i++ {
		var timeout time.Duration
		if i != 0 {
			timeout = jos.backoff.When(jos)
		}
		select {
		case <-ctx.Done():
			return
		case <-timer.After(timeout):
		}

		l.Debug("one shot job running")
		if jos.metrics {
			spanstat.Start()
		}

		jos.health.OK("running")
		err = jos.fn(ctx, jos.health)

		if jos.metrics {
			second := spanstat.End(true).Seconds()
			options.metrics.OneShotRunDuration.WithLabelValues(jos.name).Observe(second)
			spanstat.Reset()
		}
		if err == nil {
			return
		} else {
			if errors.Is(err, context.Canceled) {
				jos.health.Degraded("health degrading", err)
				l.WithError(err).Error("some error happened")
				options.metrics.JobErrorsTotal.WithLabelValues("one shot job").Inc()
			}
		}

	}
	if options.shutdowner != nil && jos.shutdownnOnError {
		options.shutdowner.Shutdown(hive.ShutdownWithError(err))
	}
}

type timerOpt func(*jobTimer)
type TimerFunc func(ctx context.Context) error
type Trigger interface {
	_trigger()
	Trigger()
}

func NewTrigger() Trigger {
	return &trigger{make(chan struct{}, 1)}
}

type trigger struct {
	c chan struct{}
}

// Trigger implements Trigger.
func (t *trigger) Trigger() {
	select {
	case t.c <- struct{}{}:
	default:
	}
}

// _trigger implements Trigger.
func (t *trigger) _trigger() {}

func WithTrigger(tr Trigger) timerOpt {
	return func(jt *jobTimer) {
		jt.trigger = tr.(*trigger)
	}
}

type jobTimer struct {
	name     string
	fn       TimerFunc
	opts     []timerOpt
	health   cell.HealthReporter
	interval time.Duration
	trigger  *trigger
}

func (jt *jobTimer) start(ctx context.Context, wg *sync.WaitGroup, scope cell.Scope, options options) {
	defer wg.Done()
	for _, opt := range jt.opts {
		opt(jt)
	}

	jt.health = cell.GetHealthReporter(scope, jt.name)
	l := options.logger.WithFields(logrus.Fields{
		"name":   jt.name,
		"module": "",
	})
	tik := time.NewTicker(jt.interval)
	defer tik.Stop()

	triggerChan := make(chan struct{})
	if jt.trigger.c != nil {
		triggerChan = jt.trigger.c
	}
	spansta := spanstat.SpanStat{}

	l.Logger.Debug("start job timer")
	for {

		select {
		case <-ctx.Done():
			return
		case <-tik.C:
		case <-triggerChan:
		}

		l.Debug("timer job finished")
		spansta.Start()
		err := jt.fn(ctx)

		tot := spansta.End(true).Total()
		options.metrics.TimerRunDuration.WithLabelValues(jt.name).Observe(tot.Seconds())
		spansta.Reset()

		if err == nil {
			jt.health.OK("job timer job is healthy")
			l.WithError(err).Error("TIMER JOB ERROR OUT")
		} else {
			if errors.Is(err, context.Canceled) {
				jt.health.Degraded("job timer job degraded", err)
				options.metrics.JobErrorsTotal.WithLabelValues(jt.name).Inc()
			}
		}

		if ctx.Err() != nil {
			return
		}
	}
}

type ObserverFunc[T any] func(ctx context.Context, event T) error
type observerOpt[T any] func(*jobObserver[T])
type jobObserver[T any] struct {
	name       string
	fn         ObserverFunc[T]
	opts       []observerOpt[T]
	observable stream.Observable[T]
	health     cell.HealthReporter
	shutdown   hive.Shutdowner
}

func Observer[T any](name string, fn ObserverFunc[T], observable stream.Observable[T], opts ...observerOpt[T]) Job {
	if fn == nil {
		panic("function could not be nil")
	}
	return &jobObserver[T]{
		name:       name,
		fn:         fn,
		opts:       opts,
		observable: observable,
	}
}

func (jo *jobObserver[T]) Start(ctx context.Context, wg *sync.WaitGroup, scope cell.Scope, options options) {
	defer wg.Done()

	for _, opt := range jo.opts {
		opt(jo)
	}

	l := options.logger.WithFields(logrus.Fields{
		"name": jo.name,
		"func": internal.FuncNameAndLocation(jo.fn),
	})
	jo.health = cell.GetHealthReporter(scope, "observal-job"+jo.name)
	tik := time.NewTicker(10 * time.Second)
	defer tik.Stop()

	jo.health.OK("primed")
	l.Logger.Debug("Observe job started")
	var msgCount uint64
	done := make(chan interface{})

	var (
		spansta = spanstat.SpanStat{}
		err     error
	)
	jo.observable.Observe(ctx, func(t T) {

		spansta.Start()
		err := jo.fn(ctx, t)

		tot := spansta.End(true).Total()
		options.metrics.TimerRunDuration.WithLabelValues(jo.name).Observe(tot.Seconds())
		spansta.Reset()

		if err != nil {
			if errors.Is(err, context.Canceled) {
				return

			}
			jo.health.Degraded("job timer job error", err)
			options.metrics.JobErrorsTotal.WithLabelValues(jo.name).Inc()
			if jo.shutdown != nil {
				jo.shutdown.Shutdown(hive.ShutdownWithError(err))
			}
			return
		}
		msgCount += 1
		select {
		case <-tik.C:
			jo.health.OK("observal job OK")
		}

	}, func(err error) {
		err = err
		close(done)
	})
	<-done
	jo.health.Stopped("observer job stopped")
	if err != nil {
		l.Debug("observal job err")
	} else {
		l.Debug("observal job ok")
	}
}
