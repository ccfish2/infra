package script

import (
	"bytes"
	"context"
	"io"

	"github.com/spf13/pflag"
)

type State struct {
	engine *Engine // the engine currently executing the script, if any

	ctx    context.Context
	cancel context.CancelFunc
	log    bytes.Buffer
	logOut io.Writer

	workdir string            // initial working directory
	pwd     string            // current working directory during execution
	env     []string          // environment list (for os/exec)
	envMap  map[string]string // environment mapping (matches env)
	stdout  string            // standard output from last 'go' command; for 'stdout' command
	stderr  string            // standard error from last 'go' command; for 'stderr' command

	Flags       *pflag.FlagSet
	DoUpdate    bool
	FileUpdates map[string]string
	RetryCount  int

	retriesRequested bool

	BreakOnError bool

	background []backgroundCmd
}

func (s *State) LogWriter() io.Writer {
	return &s.log
}

type backgroundCmd struct {
	*command
	wait WaitFunc
}

type funcCmd struct {
	usage CmdUsage
	run   func(*State, ...string) (WaitFunc, error)
}

func (c *funcCmd) Run(s *State, args ...string) (WaitFunc, error) {
	return c.run(s, args...)
}

func (c *funcCmd) Usage() *CmdUsage { return &c.usage }

func Command(usage CmdUsage, run func(*State, ...string) (WaitFunc, error)) Cmd {
	return &funcCmd{
		usage: usage,
		run:   run,
	}
}
