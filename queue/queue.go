package queue

import (
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/JackalLabs/sequoia/config"

	"github.com/jackalLabs/canine-chain/v4/x/storage/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"

	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/rs/zerolog/log"
)

// does it really need to sleep less than 0.1 second?
const minSleepDuration = time.Millisecond * 500

// NewQueue creates a new queue. There must be at least 1 thread
// For every thread, TxWorker is assigned its own wallet with an offset
// Each TxWorker uses its wallet to sign and broadcast messages
// Queue sleeps for refreshInterval after Queue distributes messages to the workers
func NewQueue(w *wallet.Wallet, config config.QueueConfig, authedClaimers []string) (*Queue, error) {
	if config.QueueThreads < 1 {
		return nil, errors.New("thread must be at least 1")
	}

	refreshInterval := time.Duration(config.QueueInterval) * time.Second
	if refreshInterval < minSleepDuration {
		return nil, errors.New("refresh interval is too short") // TODO: improve this
	}

	q := &Queue{
		wallet:          w,
		txWorkers:       make([]*TxWorker, config.QueueThreads),
		pool:            make([]*Message, 0),
		poolLock:        &sync.Mutex{},
		refreshInterval: refreshInterval,
		processed:       time.Now(),
		running:         false,
	}

	// temp tx worker to init children workers
	parent := NewTxWorker(
		q,
		0,
		config.WorkerQueueSize,
		config.TxBatchSize,
		config.MaxRetryAttempt,
		w)

	for i := 0; i < int(config.QueueThreads); i++ {
		offset := byte(i + 1)
		w, err := w.CloneWalletOffset(offset)
		if err != nil {
			// should be debugged if this happens
			return nil, errors.Join(err, errors.New("failed to create offset wallet for txWorker"))
		}

		// TODO: batch all auth claim tx into one tx
		err = q.authNewClaimer(parent, w, authedClaimers)
		if err != nil {
			return nil, err // don't want partial amount of threads
		}

		worker := NewTxWorker(
			q,
			int8(i),
			config.WorkerQueueSize,
			config.TxBatchSize,
			config.MaxRetryAttempt,
			w)

		q.txWorkers = append(q.txWorkers, worker)
	}

	return q, nil
}

func (q *Queue) Add(msg sdk.Msg) (*Message, *sync.WaitGroup) {
	var wg sync.WaitGroup

	wg.Add(1)

	m := &Message{
		msg: msg,
		wg:  &wg,
		err: nil,
	}

	q.addLast(m)

	return m, &wg
}

func (q *Queue) Listen() {
	q.running = true

	for _, w := range q.txWorkers {
		w.start()
	}
	log.Info().Msg("Queue module started")
}

func (q *Queue) Stop() {
	q.running = false
	for _, w := range q.txWorkers {
		w.stop()
	}
	log.Info().Msg("Queue module stopped")
}

func (q *Queue) addLast(msgs *Message) {
	q.poolLock.Lock()
	defer q.poolLock.Unlock()

	q.pool = append(q.pool, msgs)
}

func (q *Queue) request(count int) []*Message {
	q.poolLock.Lock()
	defer q.poolLock.Unlock()
	if len(q.pool) == 0 {
		return nil
	}

	size := min(len(q.pool), count)
	msgs := q.pool[:size]

	q.pool = q.pool[size:]
	return msgs
}

func (q *Queue) authNewClaimer(worker *TxWorker, w *wallet.Wallet, authedClaimers []string) error {
	if slices.Contains(authedClaimers, w.AccAddress()) {
		return nil
	}

	msg := types.NewMsgAddClaimer(q.wallet.AccAddress(), w.AccAddress())

	allowance := feegrant.BasicAllowance{
		SpendLimit: nil,
		Expiration: nil,
	}

	wadd, err := sdk.AccAddressFromBech32(w.AccAddress())
	if err != nil {
		panic(err)
	}

	hadd, err := sdk.AccAddressFromBech32(q.wallet.AccAddress())
	if err != nil {
		panic(err)
	}

	grantMsg, nerr := feegrant.NewMsgGrantAllowance(&allowance, wadd, hadd)
	if nerr != nil {
		panic(nerr)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	addClaimerMsg := &Message{
		msg: msg,
		wg:  &wg,
		err: nil,
	}

	feeGrantMsg := &Message{
		msg: grantMsg,
		wg:  &wg,
		err: nil,
	}

	err = worker.processBatch([]*Message{addClaimerMsg, feeGrantMsg})

	wg.Wait()
	if errs := errors.Join(err, addClaimerMsg.err, feeGrantMsg.err); errs != nil {
		return errs
	}

	return nil
}
