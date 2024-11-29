package lock

import (
	"github.com/sasha-s/go-deadlock"
)

type RWMutex struct {
	internalRWMutex
}

type Mutex struct {
	internalMutex
}

type RWMutexDebug struct {
	deadlock.RWMutex
}

type MutexDebug struct {
	deadlock.Mutex
}
