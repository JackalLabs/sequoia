package queue

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"sync"
	"time"
)

type Queue struct {
	wallet    *wallet.Wallet
	messages  []*Message
	processed time.Time
	running   bool
	interval  int64
}

type Message struct {
	msg types.Msg
	wg  *sync.WaitGroup
	err error
	res *types.TxResponse
}

func (m *Message) Error() error {
	return m.err
}

func (m *Message) Log() string {
	return m.res.RawLog
}
