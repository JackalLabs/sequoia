package cmd

import (
	"github.com/spf13/cobra"
	"sequoia/core"
)

func StartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Starts the provider",
		Run: func(cmd *cobra.Command, args []string) {

			app := core.NewApp()

			app.Start()

		},
	}
}
