package wallet

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"

	"github.com/JackalLabs/sequoia/cmd/types"
	"github.com/JackalLabs/sequoia/config"
	"github.com/spf13/cobra"
)

func WalletCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "wallet",
		Short: "Wallet subcommands",
	}

	c.AddCommand(addressCmd(), withdrawCMD())

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

			fmt.Printf("Provider Address: %s\n", wallet.AccAddress())
		},
	}
}

func withdrawCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw [to-address] [amount]",
		Short: "withdraw tokens from account",
		Args:  cobra.ExactArgs(2),
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

			c, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				panic(err)
			}

			m := bankTypes.MsgSend{
				FromAddress: wallet.AccAddress(),
				ToAddress:   args[0],
				Amount:      sdk.NewCoins(c),
			}

			data := walletTypes.NewTransactionData(
				&m,
			).WithGasAuto().WithFeeAuto()

			res, err := wallet.BroadcastTxCommit(data)
			if err != nil {
				panic(err)
			}

			if res.Code == 0 {
				fmt.Println("Withdraw successful!")
			} else {
				fmt.Println("Something went wrong, please try again.")
			}
		},
	}
}
