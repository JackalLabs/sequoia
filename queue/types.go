package queue

import (
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
)

type Queue struct {
	wallet       *wallet.Wallet
	messages     []*Message
	processed    time.Time
	running      bool
	interval     int64
	maxSizeBytes int64
	domain       string
}

type Message struct {
	msg      types.Msg
	wg       *sync.WaitGroup
	err      error
	res      *types.TxResponse
	msgIndex int
}

func (m *Message) Error() error {
	return m.err
}

func (m *Message) Log() string {
	return m.res.RawLog
}

func (m *Message) Res() *types.TxResponse {
	return m.res
}

func (m *Message) Hash() string {
	return m.res.TxHash
}

func (m *Message) Index() int {
	return m.msgIndex
}
