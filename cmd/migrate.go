package cmd

import (
	"github.com/JackalLabs/sequoia/cmd/types"
	"github.com/JackalLabs/sequoia/core"
	"github.com/spf13/cobra"
)

func JprovMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "migrate-jprov",
		Short:   "move jprov storage to sequoia",
		Long:    "jprov converts its storage to sequoia file system. This command moves that storage to sequoia data directory.",
		Example: "migrate-jprov $HOME/.jackal-storage",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			jprovRootDir := args[0]

			app := core.NewApp(home)

			app.Migrate(jprovRootDir)

			app.Start()

			return nil
		},
	}
}
