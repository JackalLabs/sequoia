package queue

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"sync"
	"time"
)

func (m *Message) Done() {
	m.wg.Done()
}

func NewQueue(w *wallet.Wallet) *Queue {
	q := &Queue{
		wallet:    w,
		messages:  make([]*Message, 0),
		processed: time.Now(),
		running:   false,
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
		time.Sleep(time.Millisecond * 333)                         // pauses for one third of a second
		if !q.processed.Add(time.Second * 10).Before(time.Now()) { // check every ten seconds
			continue
		}

		lmsg := len(q.messages)

		fmt.Printf("Queue: Posting %d messages to chain...\n", lmsg)

		l := 20
		if lmsg < l {
			l = lmsg
		}

		toProcess := q.messages[:l]
		q.messages = q.messages[l:]

		allMsgs := make([]types.Msg, len(toProcess))

		for i, process := range toProcess {
			allMsgs[i] = process.msg
		}

		data := walletTypes.NewTransactionData(
			allMsgs...,
		).WithGasAuto().WithFeeAuto()

		res, err := q.wallet.BroadcastTxCommit(data)
		for _, process := range toProcess {
			process.err = err
			process.res = res
			process.Done()
		}

		q.processed = time.Now()
	}

}
