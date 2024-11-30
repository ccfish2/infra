package cell

import (
	"context"
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	"github.com/ccfish2/infra/pkg/lock"
	"github.com/ccfish2/infra/pkg/stream"

	"golang.org/x/exp/maps"
)

type Level string

const (
	StatusUnknown Level = "Unknown"

	StatusStopped Level = "Stopped"

	StatusDegraded Level = "Degraded"

	StatusOK Level = "OK"
)

type HealthReporter interface {
	OK(status string)

	Stopped(reason string)
	Degraded(reason string, err error)
}

type Update interface {
	Level() Level
	String() string
	JSON() ([]byte, error)
	Timestamp() time.Time
}

type statusNodeReporter interface {
	setStatus(Update)
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

type StatusResult struct {
	Update
	FullModuleID FullModuleID
	Stopped      bool
}

type Status struct {
	Update

	FullModuleID FullModuleID

	Stopped     bool
	Final       Update
	LastOK      time.Time
	LastUpdated time.Time
}

func (s *Status) JSON() ([]byte, error) {
	if s.Update == nil {
		return nil, nil
	}
	return s.Update.JSON()
}

func (s *Status) Level() Level {
	if s.Update == nil {
		return StatusUnknown
	}
	return s.Update.Level()
}

func (s *Status) String() string {
	var sinceLast string
	if s.LastUpdated.IsZero() {
		sinceLast = "never"
	} else {
		sinceLast = time.Since(s.LastUpdated).String() + " ago"
	}
	return fmt.Sprintf("Status{ModuleID: %s, Level: %s, Since: %s, Message: %s}",
		s.FullModuleID, s.Level(), sinceLast, s.Update.String())
}

func NewHealthProvider() Health {
	p := &healthProvider{
		moduleStatuses: make(map[string]Status),
		byLevel:        make(map[Level]uint64),
		running:        true,
	}
	p.obs, p.emit, p.complete = stream.Multicast[Update]()

	return p
}

func (p *healthProvider) Subscribe(ctx context.Context, cb func(Update), complete func(error)) {
	p.obs.Observe(ctx, cb, complete)
}

func (p *healthProvider) processed() uint64 {
	return p.numProcessed.Load()
}

func (p *healthProvider) updateMetricsLocked(prev Update, curr Level) {
	// If an update is processed that transitions the level state of a module
	// then update the level counters.
	if prev.Level() != curr {
		p.byLevel[curr]++
		p.byLevel[prev.Level()]--
	}
}

func (p *healthProvider) process(id FullModuleID, u Update) {
	prev := func() Status {
		p.mu.Lock()
		defer p.mu.Unlock()

		t := time.Now()
		prev := p.moduleStatuses[id.String()]

		if !p.running {
			return prev
		}

		ns := Status{
			Update:      u,
			LastUpdated: t,
		}

		switch u.Level() {
		case StatusOK:
			ns.LastOK = t
		case StatusStopped:
			// If Stopped, set that module was stopped and preserve last known status.
			ns = prev
			ns.Stopped = true
			ns.Final = u
		}
		p.moduleStatuses[id.String()] = ns
		p.updateMetricsLocked(prev.Update, u.Level())
		log.WithField("status", ns.String()).Debug("Processed new health status")
		return prev
	}()
	p.numProcessed.Add(1)
	p.emit(u)
	if prev.Stopped {
		log.Warnf("module %q reported health status after being Stopped", id)
	}
}

func (p *healthProvider) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.running = false // following this, no new reporters will send.
	p.complete(nil)   // complete the observable, no new subscribers will receive further updates.
	return nil
}

var NoStatus = &StatusNode{Message: "No status reported", LastLevel: StatusUnknown}

func (p *healthProvider) forModule(moduleID FullModuleID) statusNodeReporter {
	p.mu.Lock()
	p.moduleStatuses[moduleID.String()] = Status{
		FullModuleID: moduleID,
		Update:       NoStatus,
	}
	p.byLevel[StatusUnknown]++
	p.mu.Unlock()

	return &reporter{
		moduleID: moduleID,
		process:  p.process,
	}
}
func (p *healthProvider) All() []Status {
	p.mu.RLock()
	all := maps.Values(p.moduleStatuses)
	p.mu.RUnlock()
	sort.Slice(all, func(i, j int) bool {
		return all[i].FullModuleID.String() < all[j].FullModuleID.String()
	})
	return all
}

// Get returns the latest status for a module, by module ID.
func (p *healthProvider) Get(moduleID FullModuleID) (Status, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	s, ok := p.moduleStatuses[moduleID.String()]
	if ok {
		return s, nil
	}
	return Status{}, fmt.Errorf("module %q not found", moduleID)
}

func (p *healthProvider) Stats() map[Level]uint64 {
	n := make(map[Level]uint64, len(p.byLevel))
	p.mu.Lock()
	maps.Copy(n, p.byLevel)
	p.mu.Unlock()
	return n
}

type healthProvider struct {
	mu lock.RWMutex

	running      bool
	numProcessed atomic.Uint64

	byLevel        map[Level]uint64
	moduleStatuses map[string]Status

	obs      stream.Observable[Update]
	emit     func(Update)
	complete func(error)
}

type reporter struct {
	moduleID FullModuleID
	process  func(FullModuleID, Update)
}

func (r *reporter) setStatus(u Update) {
	r.process(r.moduleID, u)
}
