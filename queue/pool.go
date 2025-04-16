package queue

import (
	"context"
	"errors"
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
	workers   []*worker
	wallet    *wallet.Wallet
	msgInput  chan *Message
	bufferIn  chan<- *Message
	bufferOut <-chan *Message
}

func NewPool(wallet *wallet.Wallet, config config.QueueConfig) (*Pool, error) {
	msgChannel := make(chan *Message)
	bufferIn := make(chan *Message)
	bufferOut := newBuffer(bufferIn)
	pool := &Pool{
		wallet:    wallet,
		msgInput:  msgChannel,
		bufferIn:  bufferIn,
		bufferOut: bufferOut,
	}

	workerWallets, err := initAuthClaimers(wallet, config.QueueThreads)
	if err != nil {
		return nil, errors.Join(errors.New("failed to initialize auth claimers"), err)
	}

	workers := createWorkers(workerWallets, config.MaxRetryAttempt)
	pool.workers = workers

	return pool, nil
}

func (p *Pool) Stop() {
	close(p.msgInput)
	close(p.bufferIn)
}

func (p *Pool) Listen() {
	for m := range p.msgInput {
		p.bufferIn <- m
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
		msgIndex: -1,
	}

	p.msgInput <- m

	return m, &wg
}

func createWorkers(workerWallets []*wallet.Wallet, maxRetryAttempt int8) []*worker {
	workers := make([]*worker, 0, len(workerWallets))
	for i, w := range workerWallets {
		worker := newWorker(int8(i), w, maxRetryAttempt)
		workers = append(workers, worker)
	}

	return workers
}

func initAuthClaimers(wallet *wallet.Wallet, count int8) (workerWallets []*wallet.Wallet, err error) {
	query := &storageTypes.QueryProvider{
		Address: wallet.AccAddress(),
	}

	cl := storageTypes.NewQueryClient(wallet.Client.GRPCConn)
	res, err := cl.Provider(context.Background(), query)
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

func newBuffer(in <-chan *Message) <-chan *Message {
	// From: https://blogtitle.github.io/go-advanced-concurrency-patterns-part-4-unlimited-buffer-channels/
	var buf []*Message
	out := make(chan *Message)

	go func() {
		defer close(out)
		for msg := range in {
			select {
			case out <- msg:
			default:
				buf = append(buf, msg)
			}
		}

		for _, v := range buf {
			out <- v
		}
	}()

	return out
}
