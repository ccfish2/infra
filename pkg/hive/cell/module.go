package cell

import (
	"regexp"
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

// Apply implements Cell.
func (m *module) Apply(container) error {
	panic("unimplemented")
}

// Info implements Cell.
func (m *module) Info(c container) Info {
	n := NewInfoNode(" " + m.id + " (" + m.title)
	for _, cell := range m.cells {
		n.Add(cell.Info(c))
	}
	return n
}

type ModuleID string
type FullModuleID []string
