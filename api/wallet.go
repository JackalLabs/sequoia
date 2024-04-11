package api

import (
	"fmt"
	"net/http"

	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
)

type WithdrawRequest struct {
	ToAddress string `json:"to_address"`
	Amount    string `json:"amount"`
}

type WithdrawResponse struct {
	Response string `json:"response"`
}

func WithdrawHandler(wallet *wallet.Wallet) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Content-Type", "application/json")

		fmt.Printf("WITHDRAWING... \n")
		fmt.Fprintf(w, "Wallet: %+v \n", wallet)

		var withdraw WithdrawRequest

		err := json.NewDecoder(req.Body).Decode(&withdraw)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		c, err := sdk.ParseCoinNormalized(withdraw.Amount)
		if err != nil {
			return
		}

		m := bankTypes.MsgSend{
			FromAddress: wallet.AccAddress(),
			ToAddress:   withdraw.ToAddress,
			Amount:      sdk.NewCoins(c),
		}

		fmt.Fprintf(w, "MsgSend: %+v \n", m)

		data := walletTypes.NewTransactionData(
			&m,
		).WithGasAuto().WithFeeAuto()

		res, err := wallet.BroadcastTxCommit(data)
		if err != nil {
			return
		}

		fmt.Fprintf(w, "RES: %+v \n", res)

		if res.Code == 0 {
			fmt.Fprintf(w, "Withdraw successful!")
		} else {
			fmt.Fprintf(w, "Something went wrong, please try again.")
		}
	}
}

func WithdrawHandlerTest() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Content-Type", "application/json")

		fmt.Printf("WITHDRAWING... \n")

		var withdraw WithdrawRequest

		err := json.NewDecoder(req.Body).Decode(&withdraw)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprintf(w, "Withdraw: %+v", withdraw)
	}
}
