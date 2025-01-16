package kvstore

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

type BackendOperations interface {
	// returns a channel which is closed when cilent connected to kvstore server
	Connected(ctx context.Context) <-chan error

	// returns a channel which is closed when client is not connected to kv store sever
	Disconnected(ctx context.Context) <-chan struct{}

	Status() (string, error)
	StatusCheckError() <-chan error
	LockPath(ctx context.Context, path string) (KVLocker, error)
	Get(ctx context.Context, key string) ([]byte, error)
	GetIfLocked(ctx context.Context, key string, lock KVLocker) ([]byte, error)
	GetPrefix(ctx context.Context, prefix string) (string, []byte, error)
	GetPrefixIfLocked(ctx context.Context, prefix string, lock KVLocker) (string, []byte, error)
	Set(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
}
type backendModule interface {
	GetName() string
	newClient(ctx context.Context, opts *ExtraOptions) (BackendOperations, chan error)
}
type KVLocker interface {
	Unlock(ctx context.Context) error
	Comparator() interface{}
}

type ClusterSizeDependentIntervalFunc func(interval time.Duration) time.Duration

type backendOption struct {
	description string
	value       string
	validate    func(value string) error
}
type backendOptions map[string]*backendOption

type ExtraOptions struct {
	DiaOption                    []grpc.DialOption
	ClusterSizeDependentInterval ClusterSizeDependentIntervalFunc
	NoLockQuorumCheck            bool
	ClusterName                  string
}
