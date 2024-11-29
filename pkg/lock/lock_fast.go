package lock

import (
	"sync"
)

type internalRWMutex struct {
	sync.RWMutex
}

func (i *internalRWMutex) UnlockIgnoreTime() {
	i.RWMutex.Unlock()
}

type internalMutex struct {
	sync.Mutex
}

func (i *internalMutex) UnlockIgnoreTime() {
	i.Mutex.Unlock()
}
