package store

import "context"

type SyncStoreBackend interface {
	Update(ctx context.Context, key string, value []byte, lease bool) error
	Delete(ctx context.Context, key string) error

	RegisterLeaseExpiredObserver(prefix string, fn func(key string))
}
