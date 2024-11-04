package cell

import (
	"context"
	"time"
)

type Level string

const (
	StatusUnknown  Level = "Unknown"
	StatusStopped  Level = "Stopped"
	StatusDegraded Level = "Degraded"
	StatusOK       Level = "OK"
)

type statusNodeReporter interface {
	setStatus(Update)
}

type Update interface {
	Level() Level
	String() string
	JSON() ([]byte, error)
	Timestamp() time.Time
}

type Status struct {
	Update

	FullModuleID FullModuleID

	Stopped     bool
	Final       Update
	LastOK      time.Time
	LastUpdated time.Time
}

type Health interface {
	All() []Status
	Get(FullModuleID) (Status, error)
	Stats() map[Level]uint64
	Stop(context.Context) error
	Subscribe(context.Context, func(Update), func(error))
	forModule(FullModuleID) statusNodeReporter
	processed() uint64
}

type HealthReporter interface {
	OK(status string)
	Stopped(reason string)
	Degraded(reason string, err error)
}
