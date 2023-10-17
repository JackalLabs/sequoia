package cmd

import (
	"github.com/JackalLabs/sequoia/core"
	"github.com/spf13/cobra"
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
