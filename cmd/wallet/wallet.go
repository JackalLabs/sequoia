package wallet

import (
	"fmt"
	"github.com/JackalLabs/sequoia/cmd/types"
	"github.com/JackalLabs/sequoia/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func WalletCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "wallet",
		Short: "Wallet subcommands",
	}

	c.AddCommand(addressCmd())

	return c
}

func addressCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "address",
		Short: "Check this providers address",
		Run: func(cmd *cobra.Command, args []string) {

			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				panic(err)
			}

			_, err = config.Init(home)
			if err != nil {
				panic(err)
			}

			wallet, err := config.InitWallet(home)
			if err != nil {
				panic(err)
			}

			log.Info().Msg(fmt.Sprintf("Provider Address: %s", wallet.AccAddress()))
		},
	}
}
