package cell

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/dig"

	"github.com/ccfish2/infra/pkg/hive/internal"
)

type invoker struct {
	funcs []namedFunc
}

type InvokerList interface {
	AppendInvoke(func() error)
}

type namedFunc struct {
	name string
	fn   any
	info dig.InvokeInfo
}

func (inv *invoker) invoke(cont container) error {
	for i, afn := range inv.funcs {

		t0 := time.Now()
		if err := cont.Invoke(afn.fn, dig.FillInvokeInfo(&inv.funcs[i].info)); err != nil {
			fmt.Printf("Invoke failed")
			return err
		}
		d := time.Since(t0)
		fmt.Printf("duration %s function %s Invokded", d, afn.name)
	}
	return nil
}

func (i *invoker) Apply(c container) error {

	invoker := func() error { return i.invoke(c) }

	return c.Invoke(func(l InvokerList) {
		l.AppendInvoke(invoker)
	})
}

func (i *invoker) Info(container) Info {
	n := NewInfoNode("")
	for _, namedFunc := range i.funcs {
		invNode := NewInfoNode(fmt.Sprintf("🛠️ %s", namedFunc.name))
		invNode.condensed = true

		var ins []string
		for _, input := range namedFunc.info.Inputs {
			ins = append(ins, internal.TrimName(input.String()))
		}
		sort.Strings(ins)
		invNode.AddLeaf("⇨ %s", strings.Join(ins, ", "))
		n.Add(invNode)
	}
	return n
}

// Cell is construccted using the invoke functions
func Invoke(funcs ...any) Cell {
	nameFuncs := []namedFunc{}
	for _, fn := range funcs {
		nameFuncs = append(nameFuncs,
			namedFunc{name: internal.FuncNameAndLocation(fn), fn: fn})
	}
	return &invoker{funcs: nameFuncs}
}
