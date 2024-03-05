package proofs

import (
	"time"

	"github.com/wealdtech/go-merkletree/v2"

	"github.com/JackalLabs/sequoia/queue"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
)

type Prover struct {
	running        bool
	wallet         *wallet.Wallet
	q              *queue.Queue
	processed      time.Time
	interval       int64
	io             FileSystem
	threads        int64
	currentThreads int64
	chunkSize      int
}

type FileSystem interface {
	DeleteFile([]byte, string, int64) error
	ProcessFiles(func([]byte, string, int64)) error
	GetFileTreeByChunk([]byte, string, int64, int, int) (*merkletree.MerkleTree, []byte, error)
}

func (p *Prover) Inc() {
	p.currentThreads++
}

func (p *Prover) Dec() {
	p.currentThreads--
}

func (p *Prover) Full() bool {
	return p.threads <= p.currentThreads
}
