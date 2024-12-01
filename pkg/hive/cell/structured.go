package cell

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/tabwriter"
	"time"

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/apimachinery/pkg/util/sets"
)

var reporterMinTimeout = time.Millisecond * 500

type children sets.Set[string]

type node struct {
	id         string
	name       string
	parentID   string
	isReporter bool
	count      int
	refs       int
	Message    string
	Error      error
	nodeUpdate
}
type nodeUpdate struct {
	Level
	Timestamp time.Time
}

type subreporterBase struct {
	sync.Mutex

	hr statusNodeReporter

	idToChildren map[string]children
	nodes        map[string]*node

	rootID  string
	stopped bool

	revision atomic.Uint64
	counter  atomic.Int32
	wakeup   chan struct{}
}

type subReporter struct {
	base            *subreporterBase
	scheduleRealize func()
	realizeSync     func()

	closeReconciler func()
	id              string
	name            string
}

const logReporterID = "reporterID"

var errReporterStopped = errors.New("reporter has been stopped")

func (s *subReporter) OK(message string) {
	if err := s.base.setStatus(s.id, StatusOK, message, nil); err != nil {
		if errors.Is(err, errReporterStopped) {
			log.WithError(err).WithField(logReporterID, s.id).Debug("could not set OK status on subreporter")
		} else {
			log.WithError(err).WithField(logReporterID, s.id).Warn("could not set OK status on subreporter")
		}
		return
	}

	s.scheduleRealize()
}

func (s *subReporter) Degraded(message string, err error) {
	if err := s.base.setStatus(s.id, StatusDegraded, message, err); err != nil {
		if errors.Is(err, errReporterStopped) {
			log.WithError(err).WithField(logReporterID, s.id).Debug("could not set degraded status on subreporter")
		} else {
			log.WithError(err).WithField(logReporterID, s.id).Warn("could not set degraded status on subreporter")
		}
		return
	}
	s.scheduleRealize()
}

func (s *subreporterBase) setStatus(id string, level Level, message string, err error) error {
	s.Lock()
	defer s.Unlock()

	if s.stopped {
		return fmt.Errorf("reporter tree %s has been stopped", id)
	}

	if _, ok := s.nodes[id]; !ok {
		return fmt.Errorf("could not set status for reporter %s: %w", id, errReporterStopped)
	}

	n := s.nodes[id]

	if n.Level == level && n.Message == message {
		n.count++
	} else {
		n.count = 1
	}

	n.Level = level
	n.Message = message
	n.Error = err
	return nil
}

func (s *subReporter) Stopped(reason string) {
	s.base.Lock()
	s.base.removeTreeLocked(s.id)
	s.base.Unlock()
	s.scheduleRealize()
}

type scope struct {
	*subReporter
}

func (s *scope) scope() *subReporter {
	return s.subReporter
}

func (s *scope) Name() string {
	return s.name
}

func (b *subreporterBase) removeRefLocked(id string) {
	if _, ok := b.nodes[id]; ok {
		if b.nodes[id].refs > 0 {
			b.nodes[id].refs--
		}
	}
}

func (s *scope) Close() {
	s.base.Lock()
	s.base.removeRefLocked(s.id)
	s.base.removeTreeLocked(s.id)
	s.base.Unlock()
	s.scheduleRealize()
}

func (r *scope) start() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		var rev uint64
		var lastUpdate time.Time
		for {
			select {
			case <-After(reporterMinTimeout):
			case <-r.base.wakeup:
			case <-ctx.Done():
			}

			if rev < r.base.revision.Load() && time.Since(lastUpdate) > reporterMinTimeout {
				rev = r.base.revision.Load()
				r.base.Lock()
				r.realizeSync()
				r.base.Unlock()
				lastUpdate = time.Now()
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()
	r.closeReconciler = cancel
}

func (s *subreporterBase) addNode(n *node) {
	if _, ok := s.idToChildren[n.parentID]; !ok {
		s.idToChildren[n.parentID] = children{}
	}
	s.idToChildren[n.parentID][n.id] = struct{}{}
	s.idToChildren[n.id] = children{}
	s.nodes[n.id] = n
}

func (s *subreporterBase) addChild(pid string, name string, isReporter bool) string {
	id := strconv.Itoa(int(s.counter.Add(1))) + "-" + name
	s.addNode(&node{
		id:       id,
		parentID: pid,
		count:    1,
		nodeUpdate: nodeUpdate{
			Level:     StatusUnknown,
			Timestamp: time.Now(),
		},
		name:       name,
		isReporter: isReporter,
		refs:       1,
	})
	return id
}

func (f FullModuleID) String() string {
	return strings.Join(f, ".")
}

func rootScope(id FullModuleID, hr statusNodeReporter) *scope {
	r := &subReporter{
		base: &subreporterBase{
			hr:           hr,
			idToChildren: map[string]children{},
			nodes:        map[string]*node{},
			wakeup:       make(chan struct{}, 16),
		},
	}
	r.id = r.base.addChild("", id.String(), false)
	r.base.rootID = r.id

	realize := func() {
		if r.base.stopped {
			return
		}
		statusTree := r.base.getStatusTreeLocked(r.base.rootID)
		if r.base.stopped {
			r.base.hr.setStatus(statusTree)
			return
		}
		if r.base.revision.Load() == 0 {
			return
		}
		r.base.hr.setStatus(statusTree)
	}

	r.scheduleRealize = func() {
		r.base.revision.Add(1)
		r.base.wakeup <- struct{}{}
	}
	r.realizeSync = realize

	return &scope{subReporter: r}
}

type StatusNode struct {
	ID              string        `json:"id"`
	LastLevel       Level         `json:"level,omitempty"`
	Name            string        `json:"name"`
	Message         string        `json:"message,omitempty"`
	UpdateTimestamp time.Time     `json:"timestamp"`
	Count           int           `json:"count"`
	SubStatuses     []*StatusNode `json:"sub_statuses,omitempty"`
	Error           string        `json:"error,omitempty"`
}

var _ Update = (*StatusNode)(nil)

func (s *StatusNode) Level() Level {
	return s.LastLevel
}

func (s *StatusNode) Timestamp() time.Time {
	return s.UpdateTimestamp
}

func (s *StatusNode) JSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

func (s *StatusNode) allOk() bool {
	return s.LastLevel == StatusOK
}

func (s *StatusNode) writeTo(w io.Writer, d int) {
	if len(s.SubStatuses) == 0 {
		since := "never"
		if !s.UpdateTimestamp.IsZero() {
			since = duration.HumanDuration(time.Since(s.UpdateTimestamp)) + " ago"
		}
		fmt.Fprintf(w, "%s%s\t%s\t%s\t%s\t(x%d)\n", strings.Repeat("\t", d), s.Name, s.LastLevel, s.Message, since, s.Count)
	} else {
		fmt.Fprintf(w, "%s%s\n", strings.Repeat("\t", d), s.Name)
		for _, ss := range s.SubStatuses {
			ss.writeTo(w, d+1)
		}
	}
}

func (s *StatusNode) StringIndent(ident int) string {
	if s == nil {
		return ""
	}
	buf := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)
	s.writeTo(w, ident)
	w.Flush()
	return buf.String()
}

func (s *StatusNode) String() string {
	return s.Message
}

func (s *subreporterBase) removeTreeLocked(rid string) {
	for child := range s.idToChildren[rid] {
		s.removeTreeLocked(child)
	}
	if _, ok := s.nodes[rid]; ok {
		pid := s.nodes[rid].parentID
		delete(s.idToChildren[pid], rid)
	}
	delete(s.idToChildren, rid)
	delete(s.nodes, rid)
}

func (s *subreporterBase) getStatusTreeLocked(nid string) *StatusNode {
	if children, ok := s.idToChildren[nid]; ok {
		rn := s.nodes[nid]
		n := &StatusNode{
			ID:              nid,
			Message:         rn.Message,
			Name:            rn.name,
			UpdateTimestamp: rn.Timestamp,
			Count:           rn.count,
		}
		if err := rn.Error; err != nil {
			n.Error = err.Error()
		}
		allok := true
		childIDs := maps.Keys(children)
		sort.Strings(childIDs)
		for _, child := range childIDs {
			cn := s.getStatusTreeLocked(child)
			if cn == nil {
				fmt.Printf("failed to get status for node %s", child)
				continue
			}
			n.SubStatuses = append(n.SubStatuses, cn)
			if !cn.allOk() {
				allok = false
			}
		}
		if rn.isReporter {
			n.LastLevel = rn.Level
		} else {
			if allok {
				n.LastLevel = StatusOK
			} else {
				n.LastLevel = StatusDegraded
			}
		}

		return n
	}
	return nil
}

func GetSubScope(parent Scope, name string) Scope {
	if parent == nil {
		return nil
	}
	return createSubScope(parent, name)
}

func (s *subreporterBase) canRemoveTreeLocked(id string) bool {
	if _, ok := s.nodes[id]; ok {
		node := s.nodes[id]
		if (node.isReporter) || s.nodes[id].refs > 0 {
			return false
		}
		for child := range s.idToChildren[id] {
			if !s.canRemoveTreeLocked(child) {
				return false
			}
		}
	}
	// If it does not exist, we assume it's ok to remove (noop).
	return true
}

// set finalizer for garbage collection
func createSubScope(parent Scope, name string) *scope {
	s := &scope{
		subReporter: getSubReporter(parent, name, false),
	}
	runtime.SetFinalizer(s, func(s *scope) {
		s.base.Lock()
		s.base.removeRefLocked(s.id)
		if s.base.canRemoveTreeLocked(s.id) {
			s.base.removeTreeLocked(s.id)
		}
		s.base.Unlock()
		runtime.SetFinalizer(s, nil)
	})
	return s
}

type Scope interface {
	Name() string
	Close()
	scope() *subReporter
}

/* as the name says*/
func flushAndClose(rs Scope, reason string) {
	rs.scope().base.Lock()
	defer rs.scope().base.Unlock()
	rs.scope().closeReconciler()
	rs.scope().realizeSync()
	rs.scope().base.stopped = true
	rs.scope().base.hr.setStatus(&StatusNode{
		ID:        rs.scope().base.rootID,
		LastLevel: StatusStopped,
		Message:   reason,
	})
	rs.scope().base.removeTreeLocked(rs.scope().id)
}

func After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func GetHealthReporter(parent Scope, name string) HealthReporter {
	if parent == nil {
		return &noopReporter{}
	}
	return getSubReporter(parent, name, true)
}

func getSubReporter(parent Scope, name string, isReporter bool) *subReporter {
	return scopeFromParent(parent, name, isReporter)
}

func scopeFromParent(parent Scope, name string, isReporter bool) *subReporter {
	r := parent.scope()
	r.base.Lock()
	defer r.base.Unlock()

	// If such a reporter already exists at this scope, we just return the same reporter
	// by recreating the subreporter.
	for cid := range r.base.idToChildren[r.id] {
		child := r.base.nodes[cid]
		if child.name == name {
			r.base.addRefLocked(cid)
			return &subReporter{
				base:            r.base,
				id:              cid,
				scheduleRealize: r.scheduleRealize,
				name:            name,
			}
		}
	}

	id := r.base.addChild(r.id, name, isReporter)

	return &subReporter{
		base:            r.base,
		id:              id,
		scheduleRealize: r.scheduleRealize,
		name:            name,
	}
}

func (b *subreporterBase) addRefLocked(id string) {
	if _, ok := b.nodes[id]; ok {
		b.nodes[id].refs++
	}
}

type noopReporter struct{}

func (s *noopReporter) OK(message string)                  {}
func (s *noopReporter) Degraded(message string, err error) {}
func (s *noopReporter) Stopped(message string)             {}
