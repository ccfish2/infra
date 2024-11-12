package resource

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	// dolphin
	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/metrics"
	"github.com/ccfish2/infra/pkg/promise"
	"github.com/ccfish2/infra/pkg/stream"
)

type options struct {
	transform   cache.TransformFunc
	sourceObj   func() k8sRuntime.Object
	indexers    cache.Indexers
	metricScope string
	name        string
	releasable  bool
}

type subscriber[T k8sRuntime.Object] struct {
	r         *resource[T]
	debugInfo string
	wq        workqueue.RateLimitingInterface
	options   eventsOpts
}

type resource[T k8sRuntime.Object] struct {
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	opts   options

	needed chan struct{}

	subscribers map[uint64]*subscriber[T]
	subId       uint64

	lw           cache.ListerWatcher
	synchronized bool // flipped to true when informer has synced

	storePromise  promise.Promise[Store[T]]
	storeResolver promise.Resolver[Store[T]]

	refsMu      sync.Mutex
	refs        uint64
	resetCtx    context.Context
	resetCancel context.CancelFunc
}

var _ Resource[corev1.Node] = &resource[corev1.Node]{}

type ResourceOption func(o *options)

type Resource[T k8sRuntime.Object] interface {
	stream.Observable[Event[T]]
}

func WithMetric(scope string) ResourceOption {
	return func(o *options) {
		o.metricScope = scope
	}
}

func WithIndexers(indexers cache.Indexers) ResourceOption {
	return func(o *options) {
		o.indexers = indexers
	}
}

func New[T k8sRuntime.Object](lc cell.Lifecycle, lw cache.ListerWatcher, opts ...ResourceOption) Resource[T] {
	r := &resource[T]{
		lw: lw,
	}
	r.opts.sourceObj = func() k8sRuntime.Object {
		var obj T
		return obj
	}
	for _, o := range opts {
		o(&r.opts)
	}
	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.reset()
	lc.Append(r)
	return r
}

func (r *resource[T]) Stop(stopCtx cell.HookContext) error {
	if r.opts.releasable {
		r.refsMu.Lock()
		defer r.refsMu.Unlock()
	}

	r.cancel()
	r.wg.Wait()
	return nil
}

func (r *resource[T]) Observe(ctx context.Context, next func(Event[T]), complete func(error)) {
	stream.FromChannel(r.Events(ctx)).Observe(ctx, next, complete)
}

type eventsOpts struct {
	rateLimiter  workqueue.RateLimiter
	errorHandler ErrorHandler
}

type EventsOpts func(*eventsOpts)
type syncWorkItem struct{}

func (syncWorkItem) isWorkItem() {}

type keyWorkItem struct {
	key Key
}

func (keyWorkItem) isWorkItem() {}

func (s *subscriber[T]) enqueueSync() {
	s.wq.Add(syncWorkItem{})
}

func (s *subscriber[T]) enqueuekey(key Key) {
	s.wq.Add(keyWorkItem{key})
}

// create a new pointer to the reconciled resource type
func (r *resource[T]) resourceName() string {
	if r.opts.name != "" {
		return r.opts.name
	}

	o := *new(T)
	sourceObj := reflect.New(reflect.TypeOf(o).Elem()).Interface().(T)

	gvk, err := apiutil.GVKForObject(sourceObj, scheme)
	if err != nil {
		return ""
	}
	return strings.ToLower(gvk.Kind)
}
func (r *resource[T]) Events(ctx context.Context, opts ...EventsOpts) <-chan Event[T] {
	_, callerFile, callerLine, _ := runtime.Caller(1)
	debugInfo := fmt.Sprintf("%T.Events() called from %s:%d", r, callerFile, callerLine)
	options := eventsOpts{
		errorHandler: AlwaysRetry,
		rateLimiter:  workqueue.DefaultControllerRateLimiter(),
	}
	for _, apply := range opts {
		apply(&options)
	}
	r.markNeeded()

	out := make(chan Event[T])
	ctx, subCancel := context.WithCancel(ctx)

	sub := &subscriber[T]{
		r:         r,
		options:   options,
		debugInfo: debugInfo,
		wq: workqueue.NewRateLimitingQueueWithConfig(options.rateLimiter,
			workqueue.RateLimitingQueueConfig{Name: r.resourceName()}),
	}

	r.wg.Add()
	go func() {
		defer r.release()
		defer r.wg.Done()
		defer close(out)

		store, err := r.storePromise.Await(ctx)
		if err != nil {
			return
		}

		r.mu.Lock()
		subId := r.subId
		r.subId++
		r.subscribers[subId] = sub

		initialKeys := store.IterKeys()
		for initialKeys.Next() {
			sub.enqueuekey(initialKeys.Key())
		}

		if r.synchronized {
			sub.enqueueSync()
		}
		r.mu.Unlock()

		sub.processLoop(ctx, out, store)

		r.mu.Lock()
		delete(r.subscribers, subId)
		r.mu.Unlock()
	}()

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		select {
		case <-r.ctx.Done():
		case <-r.resetCtx.Done():
		case <-ctx.Done():
		}
		subCancel()
		sub.wq.ShutDownWithDrain()
	}()
	return out
}

type lastKnownObject[T k8sRuntime.Object] struct {
	mu   sync.RWMutex
	objs map[Key]T
}

func (l *lastKnownObject[T]) Load(key Key) (obj T, ok bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	obj, ok = l.objs[key]
	return
}

func (l *lastKnownObject[T]) Store(key Key, obj T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.objs == nil {
		l.objs = map[Key]T{}
	}
	l.objs[key] = obj
}

func (l *lastKnownObject[T]) DeleteByUID(key Key, objToDelete T) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if obj, ok := l.objs[key]; ok {
		if getUID(obj) == getUID(objToDelete) {
			delete(l.objs, key)
		}
	}
}

func getUID(obj k8sRuntime.Object) types.UID {
	meta, err := meta.Accessor(obj)
	if err != nil {
		panic(err)
	}
	return meta.GetUID()
}

func (s *subscriber[T]) processLoop(ctx context.Context, out chan Event[T], store Store[T]) {
	// shutdown workqueue
	defer s.wq.ShutDown()

	// define doneFinalizer func(done *bool)
	doneFinalizer := func(done *bool) {
		panic(fmt.Sprintf("has a broken event handler that did not"))
	}

	var lastKnownObjects lastKnownObject[T]
	// loop is label
loop:
	for {

		workItem, shutdown := s.getWorkItem()
		if shutdown {
			break
		}

		var event Event[T]

		switch workItem := workItem.(type) {
		case syncWorkItem:
			event.Kind = Sync
		case keyWorkItem:
			obj, exists, err := store.GetByKey(workItem.key)
			if !exists || err != nil {
				deletedObject, ok := lastKnownObjects.Load(workItem.key)
				if !ok {
					s.wq.Done(workItem)
					continue loop
				}
				event.Kind = Delete
				event.Key = workItem.key
				event.Object = deletedObject
			} else {
				lastKnownObjects.Store(workItem.key, obj)
				event.Kind = Upsert
				event.Key = workItem.key
				event.Object = obj
			}
		default:
			panic(fmt.Sprintf("%T: unknown work item %T", s.r, workItem))
		}

		var eventDoneSentinel = new(bool)
		event.Done = func(err error) {
			runtime.SetFinalizer(eventDoneSentinel, nil)

			if err == nil && event.Kind == Delete {
				lastKnownObjects.DeleteByUID(event.Key, event.Object)
			}
			s.eventDone(workItem, err)
			s.r.metricEventProcessed(event.Kind, err == nil)
		}

		runtime.SetFinalizer(eventDoneSentinel, doneFinalizer)
		select {
		case out <- event:
		case <-ctx.Done():
			event.Done(nil)

			for {
				_, shutdown := s.getWorkItem()
				if shutdown {
					return
				}
			}
		}
	}

	// for loop
	//		retrieve an item from the subscribers queue getWrokItem()
	// 		and then fetch the object from the store

	// process syncWorkItem or keyWorkItem separately
}

func (r *resource[T]) metricEventProcessed(eventKind EventKind, status bool) {
	if r.opts.metricScope == "" {
		return
	}

	result := "success"
	if !status {
		result = "failed"
	}

	var action string
	switch eventKind {
	case Sync:
		return
	case Upsert:
		action = "update"
	case Delete:
		action = "delete"
	}

	metrics.KubernetesEventProcessed.WithLabelValues(r.opts.metricScope, action, result).Inc()
}

func (s *subscriber[T]) eventDone(entry workItem, err error) {
	defer s.wq.Done(entry)

	if err != nil {
		numRequeues := s.wq.NumRequeues(entry)

		var action ErrorAction
		switch entry := entry.(type) {
		case syncWorkItem:
			action = s.options.errorHandler(Key{}, numRequeues, err)
		case keyWorkItem:
			action = s.options.errorHandler(entry.key, numRequeues, err)
		default:
			panic("unhandled")
		}

		switch action {
		case ErrorActionRetry:
			s.wq.AddRateLimited(entry)
		case ErrorActionStop:
			s.wq.ShutDown()
		case ErrorActionIgnore:
			s.wq.Forget(entry)
		default:
			panic("KeyQueue: unknown action")
		}
	} else {
		s.wq.Forget(entry)
	}
}

type workItem interface {
	isWorkItem()
}

func (s *subscriber[T]) getWorkItem() (e workItem, shutdown bool) {
	var raw any
	raw, shutdown = s.wq.Get()
	if shutdown {
		return
	}
	return raw.(workItem), false
}

func (r *resource[T]) Start(cell.HookContext) error {
	r.start()
	return nil
}

func (r *resource[T]) start() {
	if r.ctx.Err() != nil {
		return
	}
	r.wg.Add(1)
	go r.startWhenNeeded()
}

// receiver
func (r *resource[T]) markNeeded() {
	if r.opts.releasable {
		r.refsMu.Lock()
		r.refs++
		r.refsMu.Unlock()
	}
	select {
	case r.needed <- struct{}{}:
	default:
	}
}
func (r *resource[T]) release() {
	if r.opts.releasable {
		return
	}

	r.refsMu.Lock()
	defer r.refsMu.Unlock()

	r.refs--
	if r.refs > 0 {
		return
	}

	r.resetCancel()
	r.wg.Wait()
	close(r.needed)

	r.reset()
	r.start()
}

func (r *resource[T]) reset() {
	r.subscribers = make(map[uint64]*subscriber[T])
	r.needed = make(chan struct{}, 1)
	r.storeResolver, r.storePromise = promise.New[Store[T]]()
	r.resetCtx, r.resetCancel = context.WithCancel(context.Background())
}

func (r *resource[T]) startWhenNeeded() {
	defer r.wg.Done()

	select {
	case <-r.ctx.Done():
		return
	case <-r.needed:
	}

	if r.ctx.Err() != nil {
		return
	}

	store, informer := r.newInformer()
	r.storeResolver.Resolve(&typedStore[T]{
		store:   store,
		release: r.release,
	})

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		informer.Run(merge(r.ctx.Done(), r.resetCtx.Done()))
	}()

	if cache.WaitForCacheSync(merge(r.ctx.Done(), r.resetCtx.Done()), informer.HasSynced) {
		r.mu.Lock()
		for _, sub := range r.subscribers {
			sub.enqueueSync()
		}
		r.synchronized = true
		r.mu.Unlock()
	}
}

func (r *resource[T]) newInformer() (cache.Indexer, cache.Controller) {
	panic("not implemented yet")
}

// merge two cases into one channel
func merge[T any](c1, c2 <-chan T) <-chan T {
	ret := make(chan T)
	go func() {
		select {
		case <-c1:
		case <-c2:
		}
		close(ret)
	}()
	return ret
}
