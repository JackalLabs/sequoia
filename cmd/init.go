package cmd

import (
	"github.com/JackalLabs/sequoia/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Init the provider on chain",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := config.Init()
			if err != nil {
				panic(err)
			}

			wallet, err := config.InitWallet()
			if err != nil {
				panic(err)
			}

			log.Info().Msg(wallet.AccAddress())

		},
	}
}
