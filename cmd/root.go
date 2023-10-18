package cmd

import (
	"github.com/JackalLabs/sequoia/cmd/wallet"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

}

func RootCmd() *cobra.Command {
	r := &cobra.Command{
		Use:   "sequoia",
		Short: "Sequoia is a fast and light-weight Jackal Storage Provider.",
	}

	r.AddCommand(StartCmd(), wallet.WalletCmd())

	return r
}

func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {

		log.Error().Err(err)
		os.Exit(1)
	}
}
