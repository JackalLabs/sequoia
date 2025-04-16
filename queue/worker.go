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
	stop            chan struct{}
	batch           []*Message
}

func newWorker(id int8, wallet *wallet.Wallet, maxRetryAttempt int8) *worker {
	return &worker{
		id:              id,
		wallet:          wallet,
		maxRetryAttempt: int8(maxRetryAttempt),
		stop:            make(chan struct{}),
	}
}

func (w *worker) Stop() {
	close(w.stop)
}

func (w *worker) start(msg <-chan *Message) {
	timer := time.NewTimer(timerDuration) // if no msg comes for 5 seconds, broadcast tx
run:
	for {
		select {
		case <-w.stop:
			timer.Stop()
			break run

		case m := <-msg:
			if len(w.batch) >= batchSize {
				w.send()
			}
			w.add(m)
			timer.Reset(timerDuration)

		case <-timer.C:
			w.send()
			timer.Stop()
		}
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
	msg := make([]types.Msg, 0, len(w.batch))
	for _, m := range w.batch {
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

	for _, m := range w.batch {
		m.res = resp
		m.err = err
		m.wg.Done()
	}
}
