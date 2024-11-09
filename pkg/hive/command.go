package hive

import (
	"github.com/ccfish2/infra/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *Hive) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hive",
		Short: "Inspect the hive",
		Run: func(cmd *cobra.Command, args []string) {
			logging.SetLogLevel(logrus.WarnLevel)
			h.PrintObjects()
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
