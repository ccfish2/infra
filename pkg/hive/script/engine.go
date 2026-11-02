package script

import (
	"time"

	"github.com/spf13/pflag"
)

type Engine struct {
	Cmds  map[string]Cmd
	Conds map[string]Cond

	// If Quiet is true, Execute deletes log prints from the previous
	// section when starting a new section.
	Quiet bool

	// RetryInterval is the minimal interval for retrying commands marked with '*'.
	// If zero, then the default retry interval is used.
	RetryInterval time.Duration

	// MaxRetryInterval is the maximum time to wait before retrying.
	MaxRetryInterval time.Duration

	// MaxRetries is the maximum number of times (excluding the first one) a
	// retrying command marked with '*' is executed before giving up. If zero,
	// the number of retries is bound only by the context.
	MaxRetries uint
}

type Cond interface {
	Eval(s *State, suffix string) (bool, error)

	Usage() *CondUsage
}

type CondUsage struct {
	Summary string // a single-line summary of when the condition is true

	Prefix bool
}

type Cmd interface {
	Run(s *State, args ...string) (WaitFunc, error)

	Usage() *CmdUsage
}

type WaitFunc func(*State) (stdout, stderr string, err error)

type command struct {
	file string
	line int

	name string // the name of the command; must be non-empty

	origArgs   []string // original arguments before pflag parsing
	args       []string // shell-expanded arguments following name
	background bool     // command should run in background (ends with a trailing &)
}

type CmdUsage struct {
	Summary string // in the style of the Name section of a Unix 'man' page, omitting the name

	Args string

	Flags func(fs *pflag.FlagSet)

	Detail []string // zero or more sentences in the style of the Description section of a Unix 'man' page

	Async bool

	RegexpArgs func(rawArgs ...string) []int

	AutocompleteArgs func(state *State, before []string, cur string) []string
	AutocompleteFlag func(state *State, args []string, flag, cur string) []string
}
