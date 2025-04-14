package queue

import (
	"time"

	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"

	"github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rs/zerolog/log"
)

// offsetWallet must be registered as authorized claimer on chain
func NewTxWorker(pool *Queue, id int8, bufferSize int16, batchSize int8, maxRetryAttempt int8, offsetWallet *wallet.Wallet) *TxWorker {
	return &TxWorker{
		pool:            pool,
		id:              id,
		wallet:          offsetWallet,
		buffer:          make([]*Message, 0, bufferSize),
		bufferSize:      bufferSize,
		batchSize:       batchSize,
		maxRetryAttempt: maxRetryAttempt,
		running:         false,
	}
}

func (t *TxWorker) Address() string {
	return t.wallet.AccAddress()
}

func (t *TxWorker) available() int {
	return int(t.bufferSize) - len(t.buffer)
}

func (t *TxWorker) getWork() {
	msgs := t.pool.request(t.available())
	t.buffer = append(t.buffer, msgs...)
}

func (t *TxWorker) getBatch() []*Message {
	size := min(len(t.buffer), int(t.batchSize))
	msgs := t.buffer[:size]
	t.buffer = t.buffer[size:]
	return msgs
}

func (t *TxWorker) processBatch(b []*Message) error {
	msgs := make([]types.Msg, len(b))
	for i, m := range b {
		msgs[i] = m.msg
	}
	data := walletTypes.NewTransactionData(msgs...).WithGasAuto().WithFeeAuto()

	var resp *types.TxResponse
	var err error

	a := 0
Retry:
	for a := 0; a < int(t.maxRetryAttempt); a++ {
		resp, err = t.wallet.BroadcastTxCommit(data)
		switch code := status.Code(err); code {
		case codes.AlreadyExists, codes.NotFound, codes.OK:
			break Retry
		}
	}

	if a == int(t.maxRetryAttempt) {
		log.Error().
			Int8("id", t.id).
			Int8("max attempt", t.maxRetryAttempt).
			Int("msg size", len(b)).
			Msg("queue worker batch msg broadcast exceeded max retry attempts")
	}

	for _, m := range b {
		m.res = resp
		m.err = err
		m.wg.Done()
	}
	return err
}

func (t *TxWorker) start() {
	t.running = true
	defer t.stop()
	for t.running {
		if len(t.buffer) == 0 {
			t.getWork() // blocks the loop
		}

		batch := t.getBatch()
		start := time.Now()
		err := t.processBatch(batch) // blocks until tx goes through or exceeds retry attempts
		if err != nil {
			log.Error().
				Err(err).
				Int8("id", t.id).
				Time("started at", start).
				Time("finished at", time.Now()).
				Int("batchSize", len(batch)).
				Msg("error while processing messages")
		}
	}
}

func (t *TxWorker) stop() {
	t.running = false
}
