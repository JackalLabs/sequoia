package queue

import (
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
)

func NewTxWorker(id int8, offsetWallet *wallet.Wallet) *TxWorker {
	return &TxWorker{
		id:     id,
		wallet: offsetWallet,
		msg:    nil,
	}
}

func (t *TxWorker) Address() string {
	return t.wallet.AccAddress()
}

func (t *TxWorker) busy() bool {
	return t.msg != nil
}

func (t *TxWorker) assign(msg *Message) {
	if msg == nil {
		t.msg = msg
	}
}

func (t *TxWorker) broadCast() {
	data := walletTypes.NewTransactionData(
		t.msg.msg,
	).WithGasAuto().WithFeeAuto()

	t.msg.res, t.msg.err = t.wallet.BroadcastTxCommit(data)
	t.msg.wg.Done()
	t.msg = nil
}
