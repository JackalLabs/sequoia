package queue

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/JackalLabs/sequoia/config"
	storageTypes "github.com/jackalLabs/canine-chain/v5/x/storage/types"

	"github.com/cosmos/cosmos-sdk/types"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/rs/zerolog/log"
)

func calculateTransactionSize(messages []types.Msg) (int64, error) {
	if len(messages) == 0 {
		return 0, nil
	}

	// Estimate transaction size based on message types and content
	// This is an approximation since we can't easily get the exact transaction size
	var totalSize int64 = 0

	// Add base transaction overhead (signature, fee, gas, etc.)
	var baseOverhead int64 = 500

	for _, msg := range messages {
		// Estimate size based on message type
		switch m := msg.(type) {
		case *storageTypes.MsgPostProof:
			// Estimate size for MsgPostProof based on its fields
			size := int64(len(m.Creator)) + int64(len(m.Item)) + int64(len(m.HashList)) +
				int64(len(m.Merkle)) + int64(len(m.Owner)) + 16 // 16 bytes for Start field
			totalSize += size
		default:
			// For other message types, use a conservative estimate
			totalSize += 500 // Default estimate for unknown message types
		}
	}

	// Add some additional overhead for transaction structure
	return totalSize + baseOverhead, nil
}

func (m *Message) Done() {
	m.wg.Done()
}

func NewQueue(w *wallet.Wallet, interval uint64, maxSizeBytes int64, domain string) *Queue {
	if maxSizeBytes == 0 {
		maxSizeBytes = config.DefaultMaxSizeBytes()
	}
	q := &Queue{
		wallet:       w,
		messages:     make([]*Message, 0),
		processed:    time.Now(),
		running:      false,
		interval:     interval,
		maxSizeBytes: maxSizeBytes,
		domain:       domain,
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
			if message == nil {
				continue
			}
			if message.msg == nil {
				continue
			}
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
		time.Sleep(time.Millisecond * 100)                                                  // pauses for one third of a second
		if !q.processed.Add(time.Second * time.Duration(q.interval+2)).Before(time.Now()) { // minimum wait for 2 seconds
			continue
		}

		total := len(q.messages)
		queueSize.Set(float64(total))

		if total == 0 { // skipping this queue cycle if there is no messages to be pushed
			continue
		}

		log.Info().Msg(fmt.Sprintf("Queue: %d messages waiting to be put on-chain...", total))

		msgs := make([]types.Msg, 0)
		cutoff := 0
		for i := 0; i < total; i++ {

			msgs = append(msgs, q.messages[i].msg)

			size, err := calculateTransactionSize(msgs)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to calculate transaction size")
				break
			}

			if size > q.maxSizeBytes {
				break
			}

			cutoff = i + 1 // cutoff is now the count of messages that fit
		}

		// If nothing fits, process at least the first message
		if cutoff == 0 {
			cutoff = 1
		}

		log.Info().Msg(fmt.Sprintf("Queue: Posting %d messages to chain...", cutoff))

		toProcess := q.messages[:cutoff]
		q.messages = q.messages[cutoff:]

		allMsgs := make([]types.Msg, len(toProcess))

		for i, process := range toProcess {
			allMsgs[i] = process.msg
		}

		data := walletTypes.NewTransactionData(
			allMsgs...,
		).WithGasAuto().WithFeeAuto().WithMemo(fmt.Sprintf("Proven by %s", q.domain))

		complete := false
		var res *types.TxResponse
		var err error
		var i int
		for !complete && i < 10 {
			i++
			res, err = q.wallet.BroadcastTxAsync(data)
			if err != nil {
				if strings.Contains(err.Error(), "tx already exists in cache") {
					if data.Sequence != nil {
						data = data.WithSequence(*data.Sequence + 1)
						continue
					}
				}
				if strings.Contains(err.Error(), "mempool is full") {
					log.Info().Msg("Mempool is full, waiting for 5 minutes before trying again")
					time.Sleep(time.Minute * 5)
					continue
				}
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
			process.Done()
		}

		q.processed = time.Now()
	}
}
