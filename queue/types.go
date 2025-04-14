package queue

import (
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
)

type TxWorker struct {
	pool   *Queue
	id     int8
	wallet *wallet.Wallet
	// not concurrent safe
	buffer          []*Message
	bufferSize      int16
	batchSize       int8
	maxRetryAttempt int8
	running         bool
}

type Queue struct {
	txWorkers       []*TxWorker
	wallet          *wallet.Wallet
	refreshInterval time.Duration
	pool            []*Message
	poolLock        *sync.Mutex
	processed       time.Time
	running         bool
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
