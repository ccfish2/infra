package hive

import (
	"context"
	"log/slog"
	"time"

	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/hive/script"
)

type ScriptCmd struct {
	Name string
	Cmd  script.Cmd
}

type ScriptCmds struct {
	cell.In

	ScriptCmds []ScriptCmd `group:"script-commands"`
}

func (sc ScriptCmds) Map() map[string]script.Cmd {
	m := make(map[string]script.Cmd, len(sc.ScriptCmds))
	for _, c := range sc.ScriptCmds {
		if c.Name != "" {
			m[c.Name] = c.Cmd
		}
	}
	return m
}

const defaultScriptTimeout = time.Minute

func hiveScriptCmd(h *Hive, log *slog.Logger) script.Cmd {
	return script.Command(
		script.CmdUsage{
			Summary: "show the hive",
		},
		func(s *script.State, args ...string) (script.WaitFunc, error) {
			switch {
			// For backwards compatibility.
			case len(args) >= 1 && args[0] == "start":
				ctx, cancel := context.WithTimeout(context.Background(), defaultScriptTimeout)
				defer cancel()
				return nil, h.Start(log, ctx)
			case len(args) >= 1 && args[0] == "stop":
				ctx, cancel := context.WithTimeout(context.Background(), defaultScriptTimeout)
				defer cancel()
				return nil, h.Stop(log, ctx)
			default:
				err := h.PrintObjects(s.LogWriter(), log)
				return nil, err
			}
		},
	)
}

func hiveStartCmd(h *Hive, log *slog.Logger) script.Cmd {
	return script.Command(
		script.CmdUsage{
			Summary: "start the hive",
		},
		func(s *script.State, args ...string) (script.WaitFunc, error) {
			ctx, cancel := context.WithTimeout(context.Background(), defaultScriptTimeout)
			defer cancel()
			return nil, h.Start(log, ctx)
		},
	)
}

func hiveStopCmd(h *Hive, log *slog.Logger) script.Cmd {
	return script.Command(
		script.CmdUsage{
			Summary: "stop the hive",
		},
		func(s *script.State, args ...string) (script.WaitFunc, error) {
			ctx, cancel := context.WithTimeout(context.Background(), defaultScriptTimeout)
			defer cancel()
			return nil, h.Stop(log, ctx)
		},
	)
}
