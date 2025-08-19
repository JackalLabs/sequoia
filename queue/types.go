package queue

import (
	"errors"
	"sync"

	"github.com/cosmos/cosmos-sdk/types"
)

var ErrReachedMaxRetry = errors.New("send retry attempt reached max retry")

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
