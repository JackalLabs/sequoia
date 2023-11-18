package proofs

import (
	"time"

	"github.com/wealdtech/go-merkletree/v2"

	"github.com/JackalLabs/sequoia/queue"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
)

type Prover struct {
	running   bool
	wallet    *wallet.Wallet
	q         *queue.Queue
	processed time.Time
	interval  int64
	io        FileSystem
}

type FileSystem interface {
	DeleteFile([]byte, string, int64) error
	ProcessFiles(func([]byte, string, int64)) error
	GetFileTreeByChunk([]byte, string, int64, int) (*merkletree.MerkleTree, []byte, error)
}
