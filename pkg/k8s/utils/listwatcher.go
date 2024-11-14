package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
