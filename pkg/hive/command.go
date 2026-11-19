package hive

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/ccfish2/infra/pkg/logging"
	"github.com/fatih/color"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var runSilent = false

// Command constructs the cobra command for hive. The hive
// command can be used to inspect the dependency graph
func (h *Hive) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hive",
		Short: "Inspect the hive",
		Run: func(cmd *cobra.Command, args []string) {
			logging.SetLogLevel(logrus.WarnLevel)
			params, _ := parseRunArgs(nil)
			fmt.Printf("got parameters %v", params)
			_ = h.PrintObjects(cmd.OutOrStdout(), slog.New(slog.Default().Handler()))
		},
		TraverseChildren: false,
	}
	h.RegisterFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		&cobra.Command{
			Use:   "dot-graph",
			Short: "Output the depend graph in graphy",
			Run: func(cmd *cobra.Command, args []string) {
				h.PrintDotGraph()
			},
			TraverseChildren: false,
		})
	return cmd
}

func parseRunArgs(args []string) (interface{}, error) {
	pargs := []interface{}{}
	for i, arg := range args {
		if i == 0 {
			continue
		}

		if strings.HasPrefix(arg, "::") {
			raw := strings.TrimPrefix(arg, "::")
			var v interface{}
			err := jsoniter.Unmarshal([]byte(raw), &v)
			if err != nil {
				color.Red(L("Arguments: %s\n"), err.Error())
				return pargs, nil
			}
			pargs = append(pargs, v)
			if !runSilent {
				color.White("args[%d]: %s\n", i-1, raw)
			}
		} else if strings.HasPrefix(arg, "\\::") {
			cleaned := "::" + strings.TrimPrefix(arg, "\\::")
			pargs = append(pargs, cleaned)
			if !runSilent {
				color.White("args[%d]: %s\n", i-1, cleaned)
			}
		} else {
			pargs = append(pargs, arg)
			if !runSilent {
				color.White("args[%d]: %s\n", i-1, arg)
			}
		}
	}
	return pargs, nil
}
