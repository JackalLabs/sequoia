package proofs

import (
	"sync/atomic"
	"time"

	merkletree "github.com/wealdtech/go-merkletree/v2"

	"github.com/JackalLabs/sequoia/queue"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"

	storageTypes "github.com/jackalLabs/canine-chain/v4/x/storage/types"
)

type Prover struct {
	running        bool
	wallet         *wallet.Wallet
	q              queue.Queue
	processed      time.Time
	interval       int64
	io             FileSystem
	threads        int32
	currentThreads atomic.Int32
	chunkSize      int
	query          storageTypes.QueryClient
}

type FileSystem interface {
	DeleteFile([]byte, string, int64) error
	ProcessFiles(func([]byte, string, int64)) error
	GetFileTreeByChunk([]byte, string, int64, int, int, int64) (*merkletree.MerkleTree, []byte, error)
}

func (p *Prover) Inc() {
	p.currentThreads.Add(1)
}

func (p *Prover) Dec() {
	p.currentThreads.Add(-1)
}

func (p *Prover) Full() bool {
	return p.threads <= p.currentThreads.Load()
}
