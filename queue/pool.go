package queue

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"sync"

	"github.com/JackalLabs/sequoia/config"
	"github.com/cosmos/cosmos-sdk/types"

	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"

	storageTypes "github.com/jackalLabs/canine-chain/v4/x/storage/types"

	"github.com/rs/zerolog/log"
)

var _ Queue = &Pool{}

type Pool struct {
	workers        []*worker
	workerChannels []chan *Message // worker id should correspond to index of this
	workerRunning  *sync.WaitGroup
	wallet         *wallet.Wallet
}

func NewPool(wallet *wallet.Wallet, queryClient storageTypes.QueryClient, config config.QueueConfig) (*Pool, error) {
	workerWallets, err := initAuthClaimers(wallet, queryClient, config.QueueThreads)
	if err != nil {
		return nil, errors.Join(errors.New("failed to initialize auth claimers"), err)
	}

	workers, workerChannels, workerRunning := createWorkers(workerWallets, int(config.TxTimer), int(config.TxBatchSize), config.MaxRetryAttempt)
	if workers == nil {
		panic("no workers created")
	}
	if workerChannels == nil {
		panic("no worker channels created")
	}
	if len(workerChannels) != len(workers) {
		panic("size of workers does not match size of worker channels")
	}

	pool := &Pool{
		wallet:         wallet,
		workers:        workers,
		workerChannels: workerChannels,
		workerRunning:  workerRunning,
	}

	return pool, nil
}

func (p *Pool) Stop() {
	for _, c := range p.workerChannels {
		close(c)
	}
	p.workerRunning.Wait()
}

func (p *Pool) Listen() {
	for _, w := range p.workers {
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

	_ = p.sendToAny(m)

	return m, &wg
}

func (p *Pool) sendToAny(msg *Message) (workerId int) {
	set := make([]reflect.SelectCase, 0, len(p.workerChannels))
	for _, ch := range p.workerChannels {
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

func createWorkers(workerWallets []*wallet.Wallet, txTimer int, batchSize int, maxRetryAttempt int8) (workers []*worker, queue []chan *Message, workerRunning *sync.WaitGroup) {
	wChannels := make([]chan *Message, 0, len(workerWallets))
	for range len(workerWallets) {
		wChannels = append(wChannels, make(chan *Message))
	}

	workerRunning = &sync.WaitGroup{}
	workerRunning.Add(len(workerWallets))

	workers = make([]*worker, 0, len(workerWallets))
	for i, w := range workerWallets {
		worker := newWorker(int8(i), w, txTimer, batchSize, maxRetryAttempt, wChannels[i], workerRunning)
		workers = append(workers, worker)
	}

	return workers, wChannels, workerRunning
}

func initAuthClaimers(wallet *wallet.Wallet, queryClient storageTypes.QueryClient, count int8) (workerWallets []*wallet.Wallet, err error) {
	query := &storageTypes.QueryProvider{
		Address: wallet.AccAddress(),
	}

	res, err := queryClient.Provider(context.Background(), query)
	if err != nil {
		return nil, errors.Join(errors.New("unable to query provider auth claimers"), err)
	}
	claimers := res.Provider.AuthClaimers

	for i := range count {
		workerWallet := newOffsetWallet(wallet, int(i))
		if !slices.Contains(claimers, workerWallet.AccAddress()) {
			err := addClaimer(wallet, workerWallet)
			if err != nil {
				return nil, errors.Join(errors.New("failed to add claimer on chain"), err)
			}
		}
		workerWallets = append(workerWallets, workerWallet)
	}

	return workerWallets, nil
}

func addClaimer(main *wallet.Wallet, claimer *wallet.Wallet) error {
	msg := storageTypes.NewMsgAddClaimer(main.AccAddress(), claimer.AccAddress())
	txData := walletTypes.NewTransactionData(msg).WithFeeAuto().WithGasAuto()

	res, err := main.BroadcastTxCommit(txData)
	if err != nil {
		return errors.Join(errors.New("unable to broadcast MsgAddClaimer"), err)
	}

	log.Info().Msg(res.TxHash)

	return nil
}

func newOffsetWallet(main *wallet.Wallet, index int) *wallet.Wallet {
	w, err := main.CloneWalletOffset(byte(index + 1))
	if err != nil {
		panic(err)
	}
	return w
}
