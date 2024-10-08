package queue

import (
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/rs/zerolog/log"
)

func (m *Message) Done() {
	m.wg.Done()
}

func NewQueue(w *wallet.Wallet, interval int64) *Queue {
	q := &Queue{
		wallet:    w,
		messages:  make([]*Message, 0),
		processed: time.Now(),
		running:   false,
		interval:  interval,
	}
	return q
}

func (q *Queue) Add(msg types.Msg) (*Message, *sync.WaitGroup) {
	var wg sync.WaitGroup

	wg.Add(1)

	m := &Message{
		msg: msg,
		wg:  &wg,
		err: nil,
	}

	log.Info().Msgf("is queue messed up? %t", q.messages == nil)

	q.messages = append(q.messages, m) // adding the message to the end of the list

	return m, &wg
}

func (q *Queue) Stop() {
	q.running = false
}

func (q *Queue) Listen() {
	q.running = true
	defer log.Info().Msg("Queue module stopped")

	log.Info().Msg("Queue module started")
	for q.running {
		time.Sleep(time.Millisecond * 333)                                                // pauses for one third of a second
		if !q.processed.Add(time.Second * time.Duration(q.interval)).Before(time.Now()) { // check every ten seconds
			continue
		}

		lmsg := len(q.messages)

		if lmsg == 0 { // skipping this queue cycle if there is no messages to be pushed
			continue
		}

		log.Info().Msg(fmt.Sprintf("Queue: %d messages waiting to be put on-chain...", lmsg))

		// maxSize := 1024 * 1024 // 1mb
		maxSize := 45

		total := len(q.messages)
		queueSize.Set(float64(total))

		if total > maxSize {
			total = maxSize
		}

		log.Info().Msg(fmt.Sprintf("Queue: Posting %d messages to chain...", total))

		toProcess := q.messages[:total]
		q.messages = q.messages[total:]

		allMsgs := make([]types.Msg, len(toProcess))

		for i, process := range toProcess {
			allMsgs[i] = process.msg
		}

		data := walletTypes.NewTransactionData(
			allMsgs...,
		).WithGasAuto().WithFeeAuto()

		res, err := q.wallet.BroadcastTxCommit(data)
		if err != nil {
			log.Warn().Err(err).Msg("tx broadcast failed from queue")
		}

		for i, process := range toProcess {
			process.err = err
			process.res = res
			process.msgIndex = i
			process.Done()
		}

		q.processed = time.Now()
	}
}
