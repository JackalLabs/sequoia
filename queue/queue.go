package queue

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/types"
)

type Queue interface {
	Add(msg types.Msg) (*Message, *sync.WaitGroup)
	Listen()
	Stop()
}
