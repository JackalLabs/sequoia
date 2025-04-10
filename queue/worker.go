package queue

import (
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"

	"github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewTxWorker(id int8, bucketSize int16, msgBatchSize int8, retryAttempt int8, offsetWallet *wallet.Wallet) *TxWorker {
	return &TxWorker{
		id:            id,
		wallet:        offsetWallet,
		msgBucketSize: bucketSize,
		msgBucket:     make([]*Message, 0, bucketSize),
		msgBatchSize:  msgBatchSize,
		retryAttempt:  retryAttempt,
	}
}

func (t *TxWorker) Address() string {
	return t.wallet.AccAddress()
}

func (t *TxWorker) available() int {
	return int(t.msgBucketSize) - len(t.msgBucket)
}

func (t *TxWorker) assign(msgs []*Message) int {
	if msgs == nil {
		return 0
	}

	// take whatever it can to fill the bucket
	// required for the sanity check
	fillCount := min(int(t.msgBucketSize)-len(t.msgBucket), len(msgs))
	m := msgs[:fillCount]

	t.msgBucket = append(t.msgBucket, m...)
	return fillCount
}

func (t *TxWorker) grabNextBatch() []*Message {
	total := min(len(t.msgBucket), int(t.msgBatchSize))
	msgs := t.msgBucket[:total]
	t.msgBucket = t.msgBucket[total:]

	return msgs
}

func (t *TxWorker) broadCast() {
	batchMessage := t.grabNextBatch()

	batch := make([]types.Msg, len(batchMessage))
	for i, m := range batchMessage {
		batch[i] = m.msg
	}
	data := walletTypes.NewTransactionData(batch...).WithGasAuto().WithFeeAuto()

	var resp *types.TxResponse
	var err error
	for attempt := 0; attempt < int(t.retryAttempt); attempt++ {
		resp, err = t.wallet.BroadcastTxCommit(data)
		if err != nil {
			// retry if network is not responding
			if code := status.Code(err); code == codes.DeadlineExceeded {
				err = nil
			} else {
				break
			}
		} else {
			break
		}

	}
	for _, m := range batchMessage {
		m.wg.Done()
		m.res = resp
		m.err = err
	}
}
