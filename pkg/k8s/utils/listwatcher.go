package utils

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type listWatcherWithModifier struct {
	inner        cache.ListerWatcher
	optsModifier func(*metav1.ListOptions)
}

// List implements cache.ListerWatcher.
func (l *listWatcherWithModifier) List(options metav1.ListOptions) (runtime.Object, error) {
	panic("unimplemented")
}

// Watch implements cache.ListerWatcher.
func (l *listWatcherWithModifier) Watch(options metav1.ListOptions) (watch.Interface, error) {
	panic("unimplemented")
}

func ListerWatcherWithModifier(lw cache.ListerWatcher, optsModifier func(*metav1.ListOptions)) cache.ListerWatcher {
	return &listWatcherWithModifier{
		inner:        lw,
		optsModifier: optsModifier,
	}
}

func ListerWatcherWithModifiers(lw cache.ListerWatcher, opts ...func(*metav1.ListOptions)) cache.ListerWatcher {
	for _, opt := range opts {
		lw = ListerWatcherWithModifier(lw, opt)
	}
	return lw
}

type typedListWatcher[T k8sRuntime.Object] interface {
	List(ctx context.Context, opts metav1.ListOptions) (T, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type genListWatcher[T k8sRuntime.Object] struct {
	lw typedListWatcher[T]
}

// List implements cache.ListerWatcher.
func (g *genListWatcher[T]) List(options metav1.ListOptions) (k8sRuntime.Object, error) {
	return g.lw.List(context.Background(), options)
}

// Watch implements cache.ListerWatcher.
func (g *genListWatcher[T]) Watch(options metav1.ListOptions) (watch.Interface, error) {
	return g.lw.Watch(context.Background(), options)
}

func ListerWatcherFromTyped[T k8sRuntime.Object](lw typedListWatcher[T]) cache.ListerWatcher {
	return &genListWatcher[T]{lw: lw}
}
