package proofs

import (
	"time"

	merkletree "github.com/wealdtech/go-merkletree/v2"

	"github.com/JackalLabs/sequoia/queue"
	"github.com/JackalLabs/sequoia/rpc"
)

type Prover struct {
	running        bool
	wallet         *rpc.FailoverClient
	q              *queue.Queue
	processed      time.Time
	interval       uint64
	io             FileSystem
	threads        int16
	currentThreads int16
	chunkSize      int
	lastCount      int
}

type FileSystem interface {
	DeleteFile([]byte, string, int64) error
	ProcessFiles(func([]byte, string, int64)) error
	GetFileTreeByChunk([]byte, string, int64, int, int, int64) (*merkletree.MerkleTree, []byte, error)
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
