package queue

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/rs/zerolog/log"
)

// does it really need to sleep less than 0.1 second?
const minSleepDuration = time.Millisecond * 500

// NewQueue creates a new queue. There must be at least 1 thread
// For every thread, TxWorker is assigned its own wallet with an offset
// Each TxWorker uses its wallet to sign and broadcast messages
// Queue sleeps for refreshInterval after Queue distributes messages to the workers
func NewQueue(w *wallet.Wallet, refreshInterval time.Duration, thread int8) (*Queue, error) {
	if thread < 1 {
		return nil, errors.New("thread must be at least 1")
	}

	if refreshInterval < time.Second {
		return nil, errors.New("refresh interval cannot be shorter than 1 second")
	}

	q := &Queue{
		wallet:          w,
		txWorkers:       make([]*TxWorker, 0),
		msgPool:         make([]*Message, 0),
		refreshInterval: refreshInterval,
		processed:       time.Now(),
		running:         false,
	}

	for i := 0; i < int(thread); i++ {
		offset := byte(i + 1)
		w, err := w.CloneWalletOffset(offset)
		if err != nil {
			// should be debugged if this happens
			return nil, errors.New("failed to create offset wallet for txWorker")
		}

		worker := NewTxWorker(int8(i), w) // need defaults here
		q.txWorkers = append(q.txWorkers, worker)
	}

	return q, nil
}

func (q *Queue) Add(msg types.Msg) (*Message, *sync.WaitGroup) {
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
	defer log.Info().Msg("Queue module stopped")

	log.Info().Msg("Queue module started")

	for q.running {
		q.sleep()

		if len(q.msgPool) == 0 {
			continue
		}

		before := len(q.msgPool)
		count := q.distributeMsg()
		log.Info().Msg(fmt.Sprintf("assigned %d out of %d messages", count, before))
	}
}

func (q *Queue) Stop() {
	q.running = false
}

// fcfs
func (q *Queue) popFront() *Message {
	if len(q.msgPool) == 0 {
		return nil
	}

	msg := q.msgPool[0]
	q.msgPool = q.msgPool[1:]
	return msg
}

// assign message to free workers from msg pool
// returns number of messages assigned
func (q *Queue) distributeMsg() int {
	if len(q.msgPool) == 0 {
		return 0
	}

	count := 0
	for _, w := range q.txWorkers {

		grabbed := w.assign(q.msgPool)
		q.msgPool = q.msgPool[grabbed:]

		go w.broadCast() // we should probably give each worker its own loop instead of calling this over and over again
		count++
	}
	return count
}

// sleep until next interval
// if slept duration is too small consider increasing interval or threads
func (q *Queue) sleep() (slept time.Duration) {
	wakeUpTime := q.processed.Add(q.refreshInterval)
	sleep := wakeUpTime.Sub(time.Now())
	if sleep > minSleepDuration {
		time.Sleep(sleep)
	}

	return sleep
}

func (q *Queue) addLast(msg *Message) {
	q.msgPool = append(q.msgPool, msg)
}
