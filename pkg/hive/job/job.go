package job

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/pprof"
	"strconv"
	"sync"
	"time"

	"github.com/ccfish2/infra/pkg/hive"
	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/hive/internal"
	"github.com/ccfish2/infra/pkg/lock"
	"github.com/ccfish2/infra/pkg/stream"
	"github.com/sirupsen/logrus"
)

var Cell = cell.Module(
	"jobs",
	"Managed background goroutines and timers",
	cell.Provide(newRegistry),
)

type Registry interface {
	NewGroup(health cell.Health, opts ...groupOpt) Group
	WithLifecycle(lifecycle cell.Lifecycle) Registry
}

type registry struct {
	logger     logrus.FieldLogger
	slogger    *slog.Logger
	shutdowner hive.Shutdowner

	// Jobs that are added before the registry is started are appended here
	appLifecycle     cell.Lifecycle
	runtimeLifecycle jobLifecycle

	metrics *jobMetrics

	mu      lock.Mutex
	groups  []Group
	started bool
}

var _ cell.HookInterface = (*registry)(nil)

func newRegistry(
	logger logrus.FieldLogger,
	shutdowner hive.Shutdowner,
	lc cell.Lifecycle,
) Registry {
	r := &registry{
		logger:       logger,
		shutdowner:   shutdowner,
		appLifecycle: lc,
		slogger:      slog.New(slog.Default().Handler()),
	}
	lc.Append(r)
	return r
}

func (c *registry) Start(cell.HookContext) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return nil
	}
	c.started = true
	return nil
}

func (c *registry) Stop(ctx cell.HookContext) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.started {
		return nil
	}
	return c.runtimeLifecycle.stop(ctx, c.slogger)
}

func (c *registry) PreStopHookMarker() {}

func (c *registry) WithLifecycle(lifecycle cell.Lifecycle) Registry {
	r := &registry{
		logger:       c.logger,
		shutdowner:   c.shutdowner,
		appLifecycle: lifecycle,
	}
	lifecycle.Append(r)
	return r
}

func (c *registry) NewGroup(health cell.Health, opts ...groupOpt) Group {
	c.mu.Lock()
	defer c.mu.Unlock()

	var options options
	options.logger = c.logger
	options.shutdowner = c.shutdowner
	options.slog = c.slogger

	for _, opt := range opts {
		opt(&options)
	}

	g := &group{
		registry: c,
		options:  options,
		health:   health,
	}

	c.groups = append(c.groups, g)

	return g
}

func (c *registry) addJobs(health cell.Health, opts options, jobs ...Job) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started {
		for _, job := range jobs {
			c.appLifecycle.Append(&queuedJob{
				registry: c,
				job:      job,
				health:   health,
				options:  opts,
			})
		}
		return
	}

	for _, job := range jobs {
		c.runtimeLifecycle.insertAndStart(
			&queuedJob{
				registry:   c,
				job:        job,
				health:     health,
				options:    opts,
				runtimeJob: true,
			},
		)
	}
}

// tracks jobs started after registration.
// and removes the jobs before registration stop
type jobLifecycle struct {
	mu sync.Mutex

	// jobs is a doubly linked-list of started jobs. The head is the latest
	// started jobs, e.g. they're in the order they should be stopped.
	jobs *queuedJob

	// stopped is true if the lifecycle has been stopped and no new jobs
	// can be added.
	stopped bool
}

func (r *jobLifecycle) insertAndStart(job *queuedJob) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stopped {
		return
	}

	if r.jobs != nil {
		r.jobs.prev = job
	}
	job.next = r.jobs
	r.jobs = job
	job.Start(context.Background())
}

func (r *jobLifecycle) remove(qj *queuedJob) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.jobs == nil {
		qj.prev = nil
		qj.next = nil
		return
	}

	if qj.next != nil {
		qj.next.prev = qj.prev
	}
	if qj == r.jobs {
		r.jobs = qj.next
	} else if qj.prev != nil {
		qj.prev.next = qj.next
	}
	qj.prev = nil
	qj.next = nil
}

func (r *jobLifecycle) stop(ctx cell.HookContext, log *slog.Logger) error {
	// Collect jobs to stop and unlink them. We must stop them without holding the
	// lock as [queuedJob.Stop] will try to call [jobLifecycle.remove].
	var jobsToStop []*queuedJob
	r.mu.Lock()
	var next *queuedJob
	for job := r.jobs; job != nil; job = next {
		next = job.next
		job.next = nil
		job.prev = nil
		jobsToStop = append(jobsToStop, job)
	}
	r.jobs = nil
	r.stopped = true
	r.mu.Unlock()
	for _, job := range jobsToStop {
		job.Stop(ctx)
		if ctx.Err() != nil {
			log.Error("Stop cancelled while waiting for job to stop",
				"job",
				job.job.info())
			break
		}
	}
	return ctx.Err()
}

type queuedJob struct {
	registry   *registry
	job        Job
	health     cell.Health
	options    options
	cancel     context.CancelFunc
	done       chan struct{}
	startedAt  time.Time
	prev       *queuedJob
	next       *queuedJob
	runtimeJob bool
}

func (qj *queuedJob) Start(cell.HookContext) error {
	qj.startedAt = time.Now()
	qj.done = make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	qj.cancel = cancel
	pprof.Do(ctx, qj.options.pprofLabels, func(ctx context.Context) {
		go func() {
			defer func() {
				qj.registry.runtimeLifecycle.remove(qj)
				if qj.runtimeJob {
					qj.registry.logger.Info("Job stopped",
						"job", qj.HookInfo(),
						"duration", time.Since(qj.startedAt))
				}
				close(qj.done)
			}()
			if qj.runtimeJob {
				qj.registry.logger.Info("Job started", "job", qj.HookInfo())
			}
			qj.job.start(ctx, qj.health, qj.options)
		}()
	})
	return nil
}

func (qj *queuedJob) HookInfo() string {
	return qj.job.info()
}

// Stop implements cell.HookInterface.
func (qj *queuedJob) Stop(ctx cell.HookContext) error {
	qj.cancel()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-qj.done:
		return nil
	}
}

var _ cell.HookDescriptiveInterface = &queuedJob{}

type Group interface {
	Add(...Job)
	Scoped(name string) ScopedGroup
}

type Job interface {
	start(ctx context.Context, health cell.Health, options options)
	info() string
}

type group struct {
	registry *registry
	options  options
	health   cell.Health
}

type options struct {
	pprofLabels pprof.LabelSet
	logger      logrus.FieldLogger
	shutdowner  hive.Shutdowner
	metrics     Metrics
	slog        *slog.Logger
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

func (jg *group) Add(jobs ...Job) {
	jg.add(jg.health, jobs...)
}

func (jg *group) add(health cell.Health, jobs ...Job) {
	jg.registry.addJobs(health, jg.options, jobs...)
}

// Scoped creates a scoped group, jobs added to this scoped group will appear as a sub-scope in the health reporter
func (jg *group) Scoped(name string) ScopedGroup {
	return &scopedGroup{
		group:  jg,
		health: jg.health.NewScope(name),
	}
}

type ScopedGroup interface {
	Add(jobs ...Job)
}

type scopedGroup struct {
	group  *group
	health cell.Health
}

func (sg *scopedGroup) Add(jobs ...Job) {
	sg.group.add(sg.health, jobs...)
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

type RetryBackoff interface {
	Wait() time.Duration
}

type ConstantBackoff time.Duration

func (d ConstantBackoff) Wait() time.Duration {
	return time.Duration(d)
}

type ExponentialBackoff struct {
	Min     time.Duration
	Max     time.Duration
	current time.Duration
}

func (e *ExponentialBackoff) Wait() time.Duration {
	if e.current == 0 {
		e.current = e.Min
	} else {
		e.current *= 2
		if e.current > e.Max {
			e.current = e.Max
		}
	}
	return e.current
}

func WithRetry(times int, backoff RetryBackoff) jobOneShotOpt {
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

	health cell.Health

	// If retry > 0, retry on error x times.
	retry           int
	backoff         RetryBackoff
	shutdownOnError bool
	metrics         bool
}

func (jos *jobOneShot) start(ctx context.Context, health cell.Health, options options) {
	for _, opt := range jos.opts {
		opt(jos)
	}

	jos.health = health.NewScope("job-" + jos.name)
	defer jos.health.Close()

	l := options.slog.With(
		"name", jos.name,
		"func", internal.FuncNameAndLocation(jos.fn))

	var err error
	var timeout time.Duration
	for i := 0; jos.retry < 0 || i <= jos.retry; i++ {
		if i != 0 {
			options.logger.Debug("Delaying retry attempt",
				"backoff", timeout,
				"retry-count", i,
			)
			select {
			case <-ctx.Done():
				return
			case <-time.After(timeout):
			}
		}

		jos.health.OK("Running")
		start := time.Now()
		err = jos.fn(ctx, jos.health)

		duration := time.Since(start)
		if options.metrics != nil {
			options.metrics.OneShotRunDuration(jos.name, duration)
		}

		switch {
		case err == nil:
			jos.health.OK("Finished (" + duration.String() + ")")
			return
		case errors.Is(err, context.Canceled) || ctx.Err() != nil:
			return
		default:
			if jos.backoff != nil && (jos.retry < 0 || i < jos.retry) {
				timeout = jos.backoff.Wait()
			}
			retriesRemain := strconv.FormatInt(int64(jos.retry-i), 10)
			if jos.retry < 0 {
				retriesRemain = "<inf>"
			} else if jos.retry == 0 {
				retriesRemain = "<none>"
			}
			msg := fmt.Sprintf("Failed (duration %s, retry %d/%s in %s)", duration, i+1, retriesRemain, timeout)
			jos.health.Degraded(msg, err)
			l.Error("Failed",
				"error", err,
				"retry", i+1,
				"remaining", retriesRemain,
				"timeout", timeout,
				"duration", duration,
			)
			if options.metrics != nil {
				options.metrics.JobError(jos.name, err)
			}
		}
	}

	if options.shutdowner != nil && jos.shutdownOnError {
		options.shutdowner.Shutdown(hive.ShutdownWithError(err))
	}
}

func (jos *jobOneShot) info() string {
	return fmt.Sprintf("%s (%s)", jos.name, internal.FuncNameAndLocation(jos.fn))
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

type Trigger interface {
	_trigger()
	Trigger()
}

func NewTrigger() *trigger {
	return &trigger{
		c: make(chan struct{}, 1),
	}
}

type trigger struct {
	debounce time.Duration

	mu            sync.Mutex
	c             chan struct{}
	lastTriggered time.Time
	folds         int
	waitStart     time.Time
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

func (t *trigger) markTriggered(name string, metrics Metrics) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.lastTriggered = time.Now()
	if metrics != nil {
		metrics.TimerTriggerStats(name, t.lastTriggered.Sub(t.waitStart), t.folds)
	}
	t.folds = 0

	select {
	case <-t.c:
	default:
	}
}

type jobTimer struct {
	name string
	fn   TimerFunc
	opts []timerOpt

	health cell.Health

	interval time.Duration
	trigger  *trigger

	shutdown hive.Shutdowner
}

func (jt *jobTimer) start(ctx context.Context, health cell.Health, options options) {
	for _, opt := range jt.opts {
		opt(jt)
	}

	jt.health = health.NewScope("timer-job-" + jt.name)
	defer jt.health.Close()

	l := options.slog.With(
		"name", jt.name,
		"func", internal.FuncNameAndLocation(jt.fn))

	var tickerChan <-chan time.Time
	if jt.interval > 0 {
		ticker := time.NewTicker(jt.interval)
		defer ticker.Stop()
		tickerChan = ticker.C
	}

	var triggerChan chan struct{}
	if jt.trigger != nil {
		triggerChan = jt.trigger.c
	}

	jt.health.OK("Primed")

	for {
		select {
		case <-ctx.Done():
			return
		case <-tickerChan:
		case <-triggerChan:
		}

		l.Debug("Timer job triggered")

		if jt.trigger != nil {
			jt.trigger.markTriggered(jt.name, options.metrics)
		}

		start := time.Now()
		err := jt.fn(ctx)
		duration := time.Since(start)

		if options.metrics != nil {
			options.metrics.TimerRunDuration(jt.name, duration)
		}

		if err == nil {
			jt.health.OK("OK (" + duration.String() + ")")
			l.Debug("Timer job finished")
		} else if !errors.Is(err, context.Canceled) {
			msg := fmt.Sprintf("Timer job failed (duration %s)", duration)
			jt.health.Degraded(msg, err)
			l.Error("Timer job failed",
				"error", err,
				"duration", duration)

			if options.metrics != nil {
				options.metrics.JobError(jt.name, err)
			}
			if jt.shutdown != nil {
				jt.shutdown.Shutdown(hive.ShutdownWithError(err))
			}
		}

		if ctx.Err() != nil {
			return
		}
	}
}

func (jt *jobTimer) info() string {
	return fmt.Sprintf("%s (%s)", jt.name, internal.FuncNameAndLocation(jt.fn))
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

	health cell.Health

	observable stream.Observable[T]

	// If not nil, call the shutdowner on error
	shutdown        hive.Shutdowner
	shutdownOnError bool
}

func (jo *jobObserver[T]) start(ctx context.Context, health cell.Health, options options) {
	for _, opt := range jo.opts {
		opt(jo)
	}

	if jo.shutdownOnError && options.shutdowner != nil {
		jo.shutdown = options.shutdowner
	}

	jo.health = health.NewScope("observer-job-" + jo.name)
	defer jo.health.Close()
	reportTicker := time.NewTicker(10 * time.Second)
	defer reportTicker.Stop()

	l := options.slog.With(
		"name", jo.name,
		"func", internal.FuncNameAndLocation(jo.fn))

	jo.health.OK("Primed")
	var msgCount uint64

	done := make(chan struct{})

	var err error
	jo.observable.Observe(ctx, func(t T) {
		start := time.Now()
		err := jo.fn(ctx, t)
		duration := time.Since(start)

		if options.metrics != nil {
			options.metrics.ObserverRunDuration(jo.name, duration)
		}

		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			msg := fmt.Sprintf("Observer job failed (duration %s)", duration)
			jo.health.Degraded(msg, err)
			l.Error("Observer job errored",
				"error", err,
				"duration", duration,
			)

			if options.metrics != nil {
				options.metrics.JobError(jo.name, err)
			}
			if jo.shutdown != nil {
				jo.shutdown.Shutdown(hive.ShutdownWithError(
					err,
				))
			}
			return
		}

		msgCount++

		select {
		case <-reportTicker.C:
			jo.health.OK("OK (" + duration.String() + ") [" + strconv.FormatUint(msgCount, 10) + "]")
		default:
		}
	}, func(e error) {
		err = e
		close(done)
	})

	<-done

	if err != nil && !errors.Is(err, context.Canceled) {
		l.Error("Observer job stopped with an error", "error", err)
	}
}

func (jo *jobObserver[T]) info() string {
	return fmt.Sprintf("%s (%s)", jo.name, internal.FuncNameAndLocation(jo.fn))
}
