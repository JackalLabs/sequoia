package queue

import (
	"reflect"
	"sync"

	"github.com/JackalLabs/sequoia/config"
	"github.com/cosmos/cosmos-sdk/types"

	"github.com/desmos-labs/cosmos-go-wallet/wallet"

	storageTypes "github.com/jackalLabs/canine-chain/v4/x/storage/types"

	"github.com/rs/zerolog/log"
)

var _ Queue = &Pool{}

type Pool struct {
	root        *worker
	rootQueue   chan *Message
	offsets     []*worker
	offsetQueue []chan *Message
}

func NewPool(main *wallet.Wallet, queryClient storageTypes.QueryClient, workerWallets []*wallet.Wallet, config config.QueueConfig) (*Pool, error) {
	root, rootQueue := createWorkers([]*wallet.Wallet{main}, int(config.TxTimer), int(config.TxBatchSize), config.MaxRetryAttempt)
	workers, workerChannels := createWorkers(workerWallets, int(config.TxTimer), int(config.TxBatchSize), config.MaxRetryAttempt)

	pool := &Pool{
		root:        root[0],
		rootQueue:   rootQueue[0],
		offsets:     workers,
		offsetQueue: workerChannels,
	}

	return pool, nil
}

func (p *Pool) Stop() {
	close(p.rootQueue)
	for _, c := range p.offsetQueue {
		close(c)
	}
}

func (p *Pool) Listen() {
	go p.root.start()
	for _, w := range p.offsets {
		go w.start()
	}
}

func (p *Pool) Add(msg types.Msg) (*Message, *sync.WaitGroup) {
	var wg sync.WaitGroup
	wg.Add(1)
	m := &Message{
		msg:      msg,
		wg:       &wg,
		err:      nil,
		res:      nil,
		msgIndex: -1, // no longer relevant(?)
	}

	to := p.sendToQueue(m)
	log.Debug().Type("msg_type", m.msg).Str("worker_id", to).Msg("message is sent to a worker")

	return m, &wg
}

func (p *Pool) sendToQueue(msg *Message) (workerId string) {
	// Auth claimers can sign and broadcast MsgPostProof but
	// other messages must be signed by main wallet
	switch msg.msg.(type) {
	case *storageTypes.MsgPostProof:
		to := p.sendToOffsets(msg)
		return p.offsets[to].Id()

	default:
		p.rootQueue <- msg
		return p.root.Id() // 0 is always root
	}
}

func (p *Pool) sendToOffsets(msg *Message) int {
	set := make([]reflect.SelectCase, 0, len(p.offsetQueue))

	for _, ch := range p.offsetQueue {
		set = append(set, reflect.SelectCase{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(ch),
			Send: reflect.ValueOf(msg),
		})
	}

	// blocks until a worker is free
	to, _, _ := reflect.Select(set)
	return to
}

func createWorkers(workerWallets []*wallet.Wallet, txTimer int, batchSize int, maxRetryAttempt int8) (workers []*worker, queue []chan *Message) {
	wChannels := make([]chan *Message, 0, len(workerWallets))
	for range len(workerWallets) {
		wChannels = append(wChannels, make(chan *Message))
	}

	workers = make([]*worker, 0, len(workerWallets))
	for i, w := range workerWallets {
		worker := newWorker(int8(i), w, txTimer, batchSize, maxRetryAttempt, wChannels[i])
		workers = append(workers, worker)
	}

	return workers, wChannels
}
