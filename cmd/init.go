package cmd

import (
	"fmt"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	storageTypes "github.com/jackalLabs/canine-chain/v3/x/storage/types"
	"github.com/spf13/cobra"
	"sequoia/core"
)

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Init the provider on chain",
		Run: func(cmd *cobra.Command, args []string) {

			wallet, err := core.CreateWallet()
			if err != nil {
				panic(err)
			}

			init := storageTypes.NewMsgInitProvider(wallet.AccAddress(), "http://127.0.0.1:3333", "1000000000", "")

			data := walletTypes.NewTransactionData(
				init,
			).WithGasAuto().WithFeeAuto()

			res, err := wallet.BroadcastTxCommit(data)
			if err != nil {
				panic(err)
			}

			fmt.Println(res.TxHash)

			fmt.Println(res.RawLog)

		},
	}
}
