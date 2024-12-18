package store

import "context"

type WSMFunc func(context.Context)

type WatchStoreManager interface {
	Register(string, WSMFunc)
	Run(context.Context)
}
