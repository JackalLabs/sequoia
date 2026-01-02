package wallet

import (
	"context"
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

	c.AddCommand(addressCmd(), withdrawCMD(), balanceCMD())

	return c
}

func addressCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "address",
		Short: "Check this providers address",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			_, err = config.Init(home)
			if err != nil {
				return err
			}

			wallet, err := config.InitWallet(home)
			if err != nil {
				return err
			}

			fmt.Printf("Provider Address: %s\n", wallet.AccAddress())
			return nil
		},
	}
}

func balanceCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "balance",
		Short: "Displays the balance of the provider.",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			_, err = config.Init(home)
			if err != nil {
				return err
			}

			fc, err := config.InitWallet(home)
			if err != nil {
				return err
			}

			queryClient := bankTypes.NewQueryClient(fc.GRPCConn())

			params := &bankTypes.QueryBalanceRequest{
				Denom:   "ujkl",
				Address: fc.AccAddress(),
			}

			res, err := queryClient.Balance(context.Background(), params)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Printf("Balance: %s\n", res.Balance)
			return nil
		},
	}
}

func withdrawCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw [to-address] [amount]",
		Short: "withdraw tokens from account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			_, err = config.Init(home)
			if err != nil {
				return err
			}

			wallet, err := config.InitWallet(home)
			if err != nil {
				return err
			}

			c, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			m := bankTypes.MsgSend{
				FromAddress: wallet.AccAddress(),
				ToAddress:   args[0],
				Amount:      sdk.NewCoins(c),
			}

			data := walletTypes.NewTransactionData(
				&m,
			).WithGasAuto().WithFeeAuto()

			res, err := wallet.Wallet().BroadcastTxCommit(data)
			if err != nil {
				return err
			}

			if res.Code == 0 {
				fmt.Println("Withdraw successful!")
			} else {
				fmt.Println("Something went wrong, please try again.")
			}
			return nil
		},
	}
}
