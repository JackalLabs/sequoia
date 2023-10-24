package queue

import (
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	storageTypes "github.com/jackalLabs/canine-chain/v3/x/storage/types"
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

	q.messages = append(q.messages, m) // adding the message to the end of the list

	return m, &wg
}

func (q *Queue) Stop() {
	q.running = false
}

func (q *Queue) Listen() {
	q.running = true
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

		maxSize := 1024 * 1024 // 1mb

		total := len(q.messages)

		var size int
		for s, message := range q.messages {

			var k interface{} = message

			// nolint:all
			switch k.(type) {
			case storageTypes.MsgPostProof:
				mpp := k.(storageTypes.MsgPostProof)
				size += mpp.Size()
			}

			if size > maxSize {
				total = s
			}

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
			log.Info().Msg(fmt.Sprintf("Failed to post from queue: %s", err.Error()))
		}
		for _, process := range toProcess {
			process.err = err
			process.res = res
			process.Done()
		}

		q.processed = time.Now()
	}
}
