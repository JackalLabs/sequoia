package queue

import (
	"time"

	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: change these to a config
const batchSize = 42
const timerDuration = time.Second * 5

type worker struct {
	id              int8
	wallet          *wallet.Wallet // offset wallet
	maxRetryAttempt int8
	msgIn           <-chan *Message // worker stop if this closes
	batch           []*Message
}

func newWorker(id int8, wallet *wallet.Wallet, maxRetryAttempt int8, msgIn <-chan *Message) *worker {
	return &worker{
		id:              id,
		wallet:          wallet,
		maxRetryAttempt: int8(maxRetryAttempt),
		msgIn:           msgIn,
	}
}

func (w *worker) start() {
	timer := time.NewTimer(timerDuration) // if no msg comes for 5 seconds, broadcast tx
run:
	for {
		select {
		case m, ok := <-w.msgIn:
			if !ok { // pool closed the channel, stop worker
				break run
			}
			if len(w.batch) >= batchSize {
				w.send()
				timer.Stop()
			}
			w.add(m)
			timer.Reset(timerDuration)

		case <-timer.C:
			if len(w.batch) > 0 {
				w.send()
			}
			timer.Stop()
		}
	}

	log.Info().Int8("id", w.id).Msg("queue worker stopped")
	if len(w.batch) > 0 { // send the remaining messages
		log.Info().Int8("id", w.id).Int("messages", len(w.batch)).Msg("sending remaining messages")
		w.send()
	}
}

func (w *worker) add(msg *Message) {
	if w.batch == nil {
		w.batch = []*Message{msg}
		return
	}

	w.batch = append(w.batch, msg)
}

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
	for ; a < int(w.maxRetryAttempt); a++ {
		resp, err = w.wallet.BroadcastTxCommit(txData)
		switch code := status.Code(err); code {
		case codes.AlreadyExists, codes.NotFound, codes.OK:
			break retry
		}
		// TODO: change this to a config
		time.Sleep(time.Second) // sleep for a bit before retrying
	}

	if a == int(w.maxRetryAttempt) {
		log.Error().
			Int8("id", w.id).
			Int8("max attempt", w.maxRetryAttempt).
			Int("msg size", len(msg)).
			Err(err).
			Msg("queue worker batch msg broadcast exceeded max retry attempts")
	}

	for _, m := range toProcess {
		m.res = resp
		m.err = err
		m.wg.Done()
	}
}
