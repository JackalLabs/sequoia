package queue

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	storageTypes "github.com/jackalLabs/canine-chain/v4/x/storage/types"

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

	m := &Message{
		msg: msg,
		wg:  &wg,
		err: nil,
	}

	proofMessage, ok := msg.(*storageTypes.MsgPostProof)
	if ok {
		for _, message := range q.messages {
			queueMessage, ok := message.msg.(*storageTypes.MsgPostProof)
			if !ok {
				continue
			}
			if bytes.Equal(
				queueMessage.Merkle, proofMessage.Merkle) &&
				queueMessage.Start == proofMessage.Start &&
				queueMessage.Owner == proofMessage.Owner {
				m.msgIndex = -1
				return m, &wg
			}
		}
	}

	wg.Add(1)

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

		complete := false
		var res *types.TxResponse
		var err error
		var i int
		for !complete && i < 10 {
			i++
			res, err = q.wallet.BroadcastTxCommit(data)
			if err != nil {
				log.Warn().Err(err).Msg("tx broadcast failed from queue")
				continue
			}

			if res != nil {
				if res.Code != 0 {
					if strings.Contains(res.RawLog, "account sequence mismatch") {
						if data.Sequence != nil {
							data = data.WithSequence(*data.Sequence + 1)
							continue
						}
					}
				}
				complete = true
			} else {
				log.Warn().Msg("response is nil")
				continue
			}
		}

		if !complete {
			err = errors.New("could not complete broadcast in 10 loops")
		}

		for i, process := range toProcess {
			process.err = err
			process.res = res
			process.msgIndex = i
			log.Debug().
				Bool("res_nil", process.res == nil).
				Bool("err_nil", process.err == nil).
				Int("msg_index", i).
				Msg("Process state before Done()")
			process.Done()
		}

		q.processed = time.Now()
	}
}
