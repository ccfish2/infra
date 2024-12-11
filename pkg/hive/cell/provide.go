package cell

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ccfish2/infra/pkg/hive/internal"
	"go.uber.org/dig"
)

type provider struct {
	ctors  []any
	infos  []dig.ProvideInfo
	export bool
}

func Provide(ctors ...any) Cell {
	return &provider{ctors: ctors, export: true}
}

func (p *provider) Apply(c container) error {
	p.infos = make([]dig.ProvideInfo, len(p.ctors))
	for i, ctor := range p.ctors {
		if err := c.Provide(ctor, dig.Export(p.export), dig.FillProvideInfo(&p.infos[i])); err != nil {
			return err
		}
	}
	return nil
}

func (p *provider) Info(container) Info {
	n := &InfoNode{}
	for i, ctor := range p.ctors {
		info := p.infos[i]
		privateSymbol := ""
		if !p.export {
			privateSymbol = "🔒️"
		}

		ctorNode := NewInfoNode(fmt.Sprintf("🚧%s %s", privateSymbol, internal.FuncNameAndLocation(ctor)))
		ctorNode.condensed = true

		var ins, outs []string
		for _, input := range info.Inputs {
			ins = append(ins, internal.TrimName(input.String()))
		}
		sort.Strings(ins)
		for _, output := range info.Outputs {
			outs = append(outs, internal.TrimName(output.String()))
		}
		sort.Strings(outs)
		if len(ins) > 0 {
			ctorNode.AddLeaf("⇨ %s", strings.Join(ins, ", "))
		}
		ctorNode.AddLeaf("⇦ %s", strings.Join(outs, ", "))
		n.Add(ctorNode)
	}
	return n
}

// similar as private variable or methods within OOP
// the cell is only visible wihtn the module it was created as well as its nested object
func ProvidePrivate(ctors ...any) Cell {
	return &provider{ctors: ctors, export: false}
}
