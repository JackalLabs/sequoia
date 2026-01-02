package queue

import (
	"sync"
	"time"

	"github.com/JackalLabs/sequoia/rpc"
	"github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/time/rate"
)

type Queue struct {
	wallet       *rpc.FailoverClient
	messages     []*Message
	processed    time.Time
	running      bool
	interval     uint64
	maxSizeBytes int64
	domain       string
	// rate limiting via token bucket
	limiter *rate.Limiter
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
