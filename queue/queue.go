package queue

import (

	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"

	storageTypes "github.com/jackalLabs/canine-chain/v4/x/storage/types"

	"github.com/cosmos/cosmos-sdk/types"
)

type Queue interface {
	Add(msg types.Msg) (*Message, *sync.WaitGroup)
	Listen()
	Stop()
}
