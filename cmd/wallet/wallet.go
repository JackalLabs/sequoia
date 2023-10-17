package wallet

import (
	"fmt"
	"github.com/JackalLabs/sequoia/config"
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

			wallet, err := config.InitWallet()
			if err != nil {
				panic(err)
			}

			fmt.Printf("Provider Address: %s\n", wallet.AccAddress())
		},
	}
}
