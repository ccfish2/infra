package allocator

import (
	"context"
	"time"
)

type AllocatorProvider interface {
	Init(context.Context)
}

type NodeEventHandler interface {
	Resync(context.Context, time.Time)
}
