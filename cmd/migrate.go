package cmd

import (
	"github.com/JackalLabs/sequoia/cmd/types"
	"github.com/JackalLabs/sequoia/core"
	"github.com/spf13/cobra"
)

func JprovMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "migrate-jprov",
		Short:   "migrate jprov storage to sequoia",
		Long:    "Migrate from jackal-provider storage to sequoia. It does not move files, it creates a copy to sequoia ipfs database.",
		Example: "migrate-jprov $HOME/.jackal-storage",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			jprovRootDir := args[0]

			app := core.NewV3App(home)

			app.Migrate(jprovRootDir)

			//app.Start()

			return nil
		},
	}
}
