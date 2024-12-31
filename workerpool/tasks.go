package workerpool

import (
	"context"
	"fmt"
)

type Task interface {
	fmt.Stringer
	Err() error
}

type task struct {
	id  string
	run func(context.Context) error
}

type taskResult struct {
	id  string
	err error
}

// Err implements Task.
func (t *taskResult) Err() error {
	return t.err
}

// String implements Task.
func (t *taskResult) String() string {
	return t.id
}

var _ Task = (*taskResult)(nil)
