package proofs

import (
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/dgraph-io/badger/v4"
	"sequoia/queue"
	"time"
)

type Prover struct {
	running   bool
	wallet    *wallet.Wallet
	db        *badger.DB
	q         *queue.Queue
	processed time.Time
	locked    bool
}
