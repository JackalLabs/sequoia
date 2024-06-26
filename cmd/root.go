package cmd

import (
	"fmt"
	"os"

	"github.com/JackalLabs/sequoia/cmd/types"
	"github.com/JackalLabs/sequoia/cmd/wallet"
	"github.com/JackalLabs/sequoia/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Level(zerolog.InfoLevel)
}

func InitCmd() *cobra.Command {
	r := &cobra.Command{
		Use:   "init",
		Short: "initializes sequoias config folder",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			_, err = config.Init(home)
			if err != nil {
				return err
			}
			_, err = config.InitWallet(home)
			if err != nil {
				return err
			}

			fmt.Println("done!")

			return nil
		},
	}

	return r
}

func VersionCmd() *cobra.Command {
	r := &cobra.Command{
		Use:   "version",
		Short: "checks the version of sequoia",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Version: %s\nCommit: %s\n", config.Version(), config.Commit())

			return nil
		},
	}

	return r
}

func RootCmd() *cobra.Command {
	r := &cobra.Command{
		Use:   "sequoia",
		Short: "Sequoia is a fast and light-weight Jackal Storage Provider.",
	}

	r.PersistentFlags().String(types.FlagHome, types.DefaultHome, "sets the home directory for sequoia")

	r.AddCommand(StartCmd(), wallet.WalletCmd(), InitCmd(), VersionCmd())

	return r
}

func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {

		log.Error().Err(err)
		os.Exit(1)
	}
}
