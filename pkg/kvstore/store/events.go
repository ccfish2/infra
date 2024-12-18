package store

import "sync"

type EventType int

const (
	EventTypeCreate EventType = iota
	EventTypeUpdate
	EventTypeModify
	EventTypeDelete
)

func (e EventType) string() string {
	switch e {
	default:
		return "unknown"
	case EventTypeCreate:
		return "create"
	}
}

type KeyValueEvent struct {
	Typ   EventType
	Key   string
	Value []byte
}
type EventChan chan KeyValueEvent
type stopChan chan struct{}
type watcher struct {
	Events      EventChan
	Prefix      string
	stopChan    stopChan
	stopOnce    sync.Once
	stopWatcher sync.WaitGroup
}
