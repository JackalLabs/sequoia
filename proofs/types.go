package proofs

import (
	"time"

	"github.com/JackalLabs/sequoia/queue"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/dgraph-io/badger/v4"
)

type Prover struct {
	running   bool
	wallet    *wallet.Wallet
	db        *badger.DB
	q         *queue.Queue
	processed time.Time
	locked    bool
	interval  int64
}
