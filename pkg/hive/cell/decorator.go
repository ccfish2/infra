package cell

import (
	"fmt"

	"github.com/ccfish2/infra/pkg/hive/internal"
)

type decorator struct {
	decorator any
	cells     []Cell
}

// Apply implements Cell.
func (d *decorator) Apply(c container, root rootContainer) error {
	scope := c.Scope(fmt.Sprintf("(decorate %s)", internal.PrettyType(d.decorator)))
	if err := scope.Decorate(d.decorator); err != nil {
		return err
	}

	for _, cell := range d.cells {
		if err := cell.Apply(scope, root); err != nil {
			return err
		}
	}
	return nil
}

// Info implements Cell.
func (d *decorator) Info(c container) Info {
	n := NewInfoNode(fmt.Sprintf(" %s:%s", internal.FuncNameAndLocation(d.decorator), internal.PrettyType(d.decorator)))
	for _, cell := range d.cells {
		n.Add(cell.Info(c))
	}
	return n
}

// Decorate receives functions and cells and returns Decorator cells
func Decorate(dtor any, cells ...Cell) Cell {
	return &decorator{
		decorator: dtor,
		cells:     cells,
	}
}

// DecorateAll takes a decorate function and apply it globally
func DecorateAll(dtor any) Cell {
	return &allDecorator{dtor}
}

type allDecorator struct {
	decorator any
}

func (d *allDecorator) Apply(_ container, rc rootContainer) error {
	return rc.Decorate(d.decorator)
}

func (d *allDecorator) Info(_ container) Info {
	n := NewInfoNode(fmt.Sprintf("🔀* %s: %s", internal.FuncNameAndLocation(d.decorator), internal.PrettyType(d.decorator)))
	return n
}
