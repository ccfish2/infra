package cell

import (
	"regexp"
	"slices"

	"github.com/sirupsen/logrus"
	"go.uber.org/dig"

	//myself
	"github.com/ccfish2/infra/pkg/logging/logfields"
)

func Module(id, title string, cells ...Cell) Cell {
	validateIDAndTitle(id, title)
	return &module{id, title, cells}
}

var (
	idRegex    = regexp.MustCompile(`^[a-z][a-z0-9_\-]{1,30}$`)
	titleRegex = regexp.MustCompile(`^[a-zA-Z0-9_\- ]{1,80}$`)
)

func validateIDAndTitle(id, title string) {
	if !idRegex.MatchString(id) {
		panic("Module ID format is not correct")
	}
	if !titleRegex.MatchString(title) {
		panic("Title format is not correct")
	}
}

type module struct {
	id    string
	title string
	cells []Cell
}

func (m *module) moduleID() ModuleID {
	return ModuleID(m.id)
}

func (f FullModuleID) append(m ModuleID) FullModuleID {
	return append(slices.Clone(f), string(m))
}

func (m *module) fullModuleID(parent FullModuleID) FullModuleID {
	return parent.append(m.moduleID())
}

type reporterHooks struct {
	rootScope *scope
}

func (r *reporterHooks) Start(ctx HookContext) error {
	r.rootScope.start()
	return nil
}

func (r *reporterHooks) Stop(ctx HookContext) error {
	flushAndClose(r.rootScope, "Hive shutting down")
	return nil
}

func createStructedScope(id FullModuleID, p Health, lc Lifecycle) Scope {
	rs := rootScope(id, p.forModule(id))
	lc.Append(&reporterHooks{rootScope: rs})
	return rs
}

// Apply implements Cell.
func (m *module) Apply(c container) error {
	scope := c.Scope(m.id)

	if err := scope.Provide(m.moduleID); err != nil {
		return err
	}

	if err := scope.Decorate(m.fullModuleID); err != nil {
		return err
	}

	if err := scope.Provide(createStructedScope, dig.Export(false)); err != nil {
		return nil
	}
	if err := scope.Decorate(m.lifecycle); err != nil {
		return err
	}

	if err := scope.Decorate(m.logger); err != nil {
		return err
	}

	for _, cell := range m.cells {
		if err := cell.Apply(scope); err != nil {
			return err
		}
	}
	return nil
}

func (m *module) logger(log logrus.FieldLogger) logrus.FieldLogger {
	return log.WithField(logfields.LogSubsys, m.id)
}

// Info implements Cell.
func (m *module) Info(c container) Info {
	n := NewInfoNode(" " + m.id + " (" + m.title)
	for _, cell := range m.cells {
		n.Add(cell.Info(c))
	}
	return n
}

func (m *module) lifecycle(lc Lifecycle, fullID FullModuleID) Lifecycle {
	switch lc := lc.(type) {
	case *DefaultLifecycle:
		return &augmentedLifecycle{
			lc,
			fullID,
		}
	case *augmentedLifecycle:
		return &augmentedLifecycle{
			lc.DefaultLifecycle,
			fullID,
		}
	default:
		return lc
	}
}

type ModuleID string
type FullModuleID []string
