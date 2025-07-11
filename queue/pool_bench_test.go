package queue

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/JackalLabs/sequoia/config"
	"github.com/JackalLabs/sequoia/testutil"
	sequoiaWallet "github.com/JackalLabs/sequoia/wallet"

	"go.uber.org/mock/gomock"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

	storageQuery := testutil.NewFakeStorageQueryClient()

	offsetWallets, err := sequoiaWallet.CreateOffsetWallets(wallet, int(queueConfig.QueueThreads))
	if err != nil {
		b.Error(err)
	}
	pool, err := NewPool(wallet, storageQuery, offsetWallets, queueConfig)
	if err != nil {
		b.Error(err)
	}

	go pool.Listen()
	defer pool.Stop()

	msg := types.NewMsgPostProof("2", []byte(""), "", 0, []byte(""), []byte(""), 0)
	b.ResetTimer()
	for range b.N {
		pool.Add(msg)
	}
}
