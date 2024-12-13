package main

import (
	"log"

	"github.com/ccfish2/infra/pkg/hive"
	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/spf13/cobra"
)

var (
	// Hive is lazy intialization, objects only get initialized when invoke methods get called
	// or some objects depend on the objects during configuration
	Hive = hive.New(
		serverCell,
		exampleMetricCell,
		jobCell,
		helloHandler,

		cell.Invoke(func(Server) {}),
	)

	cmd = cobra.Command{
		Use: "show case how to use the framework",
		Run: func(cmd *cobra.Command, args []string) {
			if err := Hive.Run(); err != nil {
				log.Fatal("hive failed to run")
			}
		},
	}
)

func main() {
	Hive.RegisterFlags(cmd.Flags())
	cmd.AddCommand(Hive.Command())
	cmd.Execute()
}
