package queue

import (
	"bytes"
	"context"
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
	"golang.org/x/time/rate"
)

// Rate limiter defaults are provided by config.DefaultRateLimitPerTokenMs and config.DefaultRateLimitBurst

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

func NewQueue(w *wallet.Wallet, interval uint64, maxSizeBytes int64, domain string, rlCfg config.RateLimitConfig) *Queue {
	if maxSizeBytes == 0 {
		maxSizeBytes = config.DefaultMaxSizeBytes()
	}
	if rlCfg.PerTokenMs == 0 {
		rlCfg.PerTokenMs = config.DefaultRateLimitConfig().PerTokenMs
	}
	if rlCfg.Burst == 0 {
		rlCfg.Burst = config.DefaultRateLimitConfig().Burst
	}
	q := &Queue{
		wallet:       w,
		messages:     make([]*Message, 0),
		processed:    time.Now(),
		running:      false,
		interval:     interval,
		maxSizeBytes: maxSizeBytes,
		domain:       domain,
		limiter:      rate.NewLimiter(rate.Every(time.Duration(rlCfg.PerTokenMs)*time.Millisecond), rlCfg.Burst),
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
		time.Sleep(time.Millisecond * 100)                                                // pauses for one third of a second
		if !q.processed.Add(time.Second * time.Duration(q.interval)).Before(time.Now()) { // minimum wait for 2 seconds
			continue
		}

		// Update gauge and attempt a broadcast cycle
		total := len(q.messages)
		queueSize.Set(float64(total))

		if total == 0 { // skipping this queue cycle if there is no messages to be pushed
			continue
		}

		// Token-bucket rate limit: allow calling BroadcastPending at most 20 times per 6 seconds
		if !q.limiter.Allow() {
			continue
		}

		// bunch into 25 message chunks if possible
		if total < 25 { // if total is less than 25 messages, and it's been less than 10 minutes passed, skip
			if q.processed.Add(time.Minute * 10).After(time.Now()) {
				continue
			}
		}

		_, _ = q.BroadcastPending()
		q.processed = time.Now()
	}
}

// BroadcastPending selects a batch that fits within max size, broadcasts it,
// updates per-message results, and returns the number of messages processed
// along with a terminal error if the broadcast attempts all failed.
func (q *Queue) BroadcastPending() (int, error) {
	total := len(q.messages)
	log.Info().Msg(fmt.Sprintf("Queue: %d messages waiting to be put on-chain...", total))

	limit := 5000
	unconfirmedTxs, err := q.wallet.Client.RPCClient.UnconfirmedTxs(context.Background(), &limit)
	if err != nil {
		log.Error().Err(err).Msg("could not get mempool status")
		return 0, err
	}
	if unconfirmedTxs.Total > 2000 {
		log.Error().Msg("Cannot post messages when mempool is too large, waiting 30 minutes")
		time.Sleep(time.Minute * 30)
		return 0, nil
	}

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

	// Process at least the first 45 messages or the total number of messages if less than 45
	if cutoff < 45 {
		cutoff = 45
	}

	if cutoff > total {
		cutoff = total
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
	var i int
	for !complete && i < 10 {
		i++
		res, err = q.wallet.BroadcastTxCommit(data)
		if err != nil {
			if strings.Contains(err.Error(), "tx already exists in cache") {
				log.Info().Msg("TX already exists in mempool, we're going to skip it.")
				continue
			}
			if strings.Contains(err.Error(), "mempool is full") {
				log.Info().Msg("Mempool is full, waiting for 30 minutes before trying again and resetting queue")
				time.Sleep(time.Minute * 30)
				q.messages = make([]*Message, 0)
				return 0, nil
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

	return cutoff, err
}

func (q *Queue) Count() int {
	return len(q.messages)
}
