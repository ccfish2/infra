package hive

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var runSilent = false

// Command constructs the cobra command for hive. The hive
// command can be used to inspect the dependency graph
func (h *Hive) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "hive",
		Short:        "Inspect the hive",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.PrintObjects(cmd.OutOrStdout(), slog.New(slog.Default().Handler()))
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
