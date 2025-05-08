package queue

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/JackalLabs/sequoia/config"

	"go.uber.org/mock/gomock"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	ttypes "github.com/tendermint/tendermint/types"

	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
)

var queueConfig config.QueueConfig = config.QueueConfig{
	QueueInterval:   10,
	QueueThreads:    4,
	MaxRetryAttempt: 100,
	TxBatchSize:     45,
	TxTimer:         3,
}

func setupNewPool(w *wallet.Wallet, config config.QueueConfig) *Pool {
	workerWallets := make([]*wallet.Wallet, 0)
	for i := range config.QueueThreads {
		workerWallet := newOffsetWallet(w, int(i))
		workerWallets = append(workerWallets, workerWallet)
	}

	workers, queue, workerRunning := createWorkers(workerWallets, int(config.TxTimer), int(config.TxBatchSize), config.MaxRetryAttempt)
	return &Pool{
		wallet:         w,
		workers:        workers,
		workerChannels: queue,
		workerRunning:  workerRunning,
	}

}

// really rough estimate because it benchmarks the whole pool.Add pipeline
func BenchmarkPoolAdd(b *testing.B) {
	wallet, queryClient, _, rpcClient := setupWalletClient(b)
	rpcClient.
		EXPECT().
		BroadcastTxCommit(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, tx ttypes.Tx) (*sdkTypes.TxResponse, error) {
			return nil, status.Error(codes.OK, "no error")
		}).
		AnyTimes()
	wallet.Client.RPCClient = rpcClient
	baseAcc := authTypes.ProtoBaseAccount()
	err := baseAcc.SetAddress(sdkTypes.AccAddress(wallet.AccAddress()))
	if err != nil {
		b.Error(err)
	}

	any, err := codectypes.NewAnyWithValue(baseAcc)
	if err != nil {
		b.Error(err)
	}
	queryResp := &authTypes.QueryAccountResponse{
		Account: any,
	}
	queryClient.EXPECT().Account(gomock.Any(), gomock.Any()).Return(queryResp, nil).AnyTimes()

	pool := setupNewPool(wallet, queueConfig)
	go pool.Listen()
	defer pool.Stop()

	msg := types.NewMsgPostProof("2", []byte(""), "", 0, []byte(""), []byte(""), 0)
	b.ResetTimer()
	for range b.N {
		pool.Add(msg)
	}
}
