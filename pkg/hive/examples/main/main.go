package main

import (
	"fmt"
	"reflect"
)

type A struct{}
type B struct{ A *A }

func NewB(A *A) *B { return &B{A} }
func NewA() *A     { return &A{} }
func ShowB(B *B) {
	fmt.Printf("%#v", B)
}

// illustrating dynamically invoke dependencies
func main() {
	c := container{
		providers: map[int]provider{},
		byType:    map[string]int{},
		objects:   map[string]reflect.Value{},
	}

	c.Provide(NewA)
	c.Provide(NewB)
	c.Invoke(ShowB)
}

type container struct {
	nextId    int
	providers map[int]provider
	byType    map[string]int
	objects   map[string]reflect.Value
}

type provider struct {
	ctor any
	ins  []reflect.Type
	outs []reflect.Type
}

func (c *container) Provide(ctor any) {
	typ := reflect.TypeOf(ctor)
	in := make([]reflect.Type, typ.NumIn())
	for i := 0; i < typ.NumIn(); i++ {
		in[i] = typ.In(i)
	}
	out := make([]reflect.Type, typ.NumOut())
	for i := 0; i < typ.NumOut(); i++ {
		out[i] = typ.Out(i)
		c.byType[typ.Out(i).String()] = c.nextId
	}
	c.providers[c.nextId] = provider{ctor, in, out}
	c.nextId += 1
}

func (c *container) construct(nm string) reflect.Value {
	obj, ok := c.objects[nm]
	if ok {
		return obj
	}
	id := c.byType[nm]
	providr := c.providers[id]
	ctro := reflect.ValueOf((providr.ctor))
	ctroType := ctro.Type()

	args := make([]reflect.Value, ctroType.NumIn())
	for i := 0; i < ctroType.NumIn(); i++ {
		args[i] = c.construct(ctroType.In(i).String())
	}
	outs := ctro.Call(args)
	for i, out := range outs {
		t := ctroType.Out(i)
		c.objects[t.String()] = out
	}
	return c.objects[nm]
}

func (c *container) Invoke(fn any) {
	val := reflect.ValueOf(fn)
	typ := val.Type()
	args := make([]reflect.Value, typ.NumIn())
	for i := 0; i < typ.NumIn(); i++ {
		args[i] = c.construct(typ.In(i).String())
	}
	val.Call(args)
}
