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

			fmt.Println(wallet.AccAddress())

			init := storageTypes.NewMsgInitProvider(wallet.AccAddress(), "http://127.0.0.1:3333", "1000000000", "")

			data := walletTypes.NewTransactionData(
				init,
			).WithGasAuto().WithFeeAuto()

			builder, err := wallet.BuildTx(data)
			if err != nil {
				panic(err)
			}

			res, err := wallet.Client.BroadcastTxCommit(builder.GetTx())
			if err != nil {
				panic(err)
			}

			fmt.Println(res.TxHash)

			fmt.Println(res.RawLog)

		},
	}
}
