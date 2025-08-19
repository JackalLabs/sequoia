package queue

import (
	"errors"
	"time"

	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"

	"github.com/cosmos/cosmos-sdk/types"

	storageTypes "github.com/jackalLabs/canine-chain/v4/x/storage/types"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type worker struct {
	wallet          *wallet.Wallet // offset wallet
	maxRetryAttempt int8
	msgIn           <-chan *Message // worker stop if this closes
	batch           []*Message
	batchSize       int
	txTimer         int
}

func newWorker(id int8, wallet *wallet.Wallet, txTimer int, batchSize int, maxRetryAttempt int8, msgIn <-chan *Message) *worker {
	return &worker{
		wallet:          wallet,
		maxRetryAttempt: int8(maxRetryAttempt),
		msgIn:           msgIn,
		batchSize:       batchSize,
		txTimer:         txTimer,
	}
}

func (w *worker) start() {
	d := time.Duration(w.txTimer) * time.Second
	timer := time.NewTimer(d) // if no msg comes for 5 seconds, broadcast tx
run:
	for {
		select {
		case m, ok := <-w.msgIn:
			if !ok { // pool closed the channel, stop worker
				break run
			}

			w.add(m)
			if len(w.batch) >= w.batchSize {
				w.send()
			}
			timer.Reset(d)

		case <-timer.C:
			if len(w.batch) > 0 {
				w.send()
			}
			timer.Stop()
		}
	}

	log.Info().Str("worker_id", w.Id()).Msg("queue worker stopped")
	if len(w.batch) > 0 {
		log.Info().
			Str("id", w.Id()).
			Str("addr", w.wallet.AccAddress()).
			Int("count", len(w.batch)).
			Msg("worker discarded remaining messages in the queue")
	}
}

func (w *worker) add(msg *Message) {
	// broadcast message as auth claimer
	if m, ok := msg.msg.(*storageTypes.MsgPostProof); ok {
		m.Creator = w.wallet.AccAddress()
		msg.msg = m
	}
	if w.batch == nil {
		w.batch = []*Message{msg}
		return
	}

	w.batch = append(w.batch, msg)
}

// TODO: make this dep inject
func (w *worker) send() {
	toProcess := w.batch
	w.batch = nil

	msg := make([]types.Msg, 0, len(toProcess))
	for _, m := range toProcess {
		msg = append(msg, m.msg)
	}

	txData := walletTypes.NewTransactionData(msg...).WithFeeAuto().WithGasAuto()

	var resp *types.TxResponse
	var err error

	a := 0

retry:
	for ; a < int(w.maxRetryAttempt+1); a++ {
		resp, err = w.wallet.BroadcastTxCommit(txData)
		switch code := status.Code(err); code {
		case codes.AlreadyExists, codes.NotFound, codes.OK:
			break retry
		}
		time.Sleep(time.Second) // sleep for a bit before retrying
	}

	if a == int(w.maxRetryAttempt+1) {
		log.Error().
			Str("id", w.Id()).
			Int8("max attempt", w.maxRetryAttempt).
			Int("msg size", len(msg)).
			Err(err).
			Msg("queue worker batch msg broadcast exceeded max retry attempts")
		err = errors.Join(ErrReachedMaxRetry, err)
	}

	for _, m := range toProcess {
		m.res = resp
		m.err = err
		m.wg.Done()
	}
}

func (w *worker) Id() string {
	return w.wallet.AccAddress()[3:8]
}
