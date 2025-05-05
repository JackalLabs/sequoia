package queue

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/p2p"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	ttypes "github.com/tendermint/tendermint/types"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/JackalLabs/sequoia/config"
	sequoiaWallet "github.com/JackalLabs/sequoia/wallet"
	wtypes "github.com/desmos-labs/cosmos-go-wallet/types"

	"github.com/JackalLabs/sequoia/testutil/mocks"

	"github.com/jackalLabs/canine-chain/v4/x/storage/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	txTypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestWorkerSendMaxRetry(t *testing.T) {
	r := require.New(t)

	chainCfg := wtypes.ChainConfig{
		RPCAddr:       "http://localhost:26657",
		GRPCAddr:      "localhost:9090",
		GasPrice:      "0.02ujkl",
		GasAdjustment: 1.5,
		Bech32Prefix:  "jkl",
	}

	s := &config.Seed{
		SeedPhrase:     "forward service profit benefit punch catch fan chief jealous steel harvest column spell rude warm home melody hat broccoli pulse say garlic you firm",
		DerivationPath: "m/44'/118'/0'/0/0",
	}

	wallet, err := sequoiaWallet.CreateWallet(s.SeedPhrase, s.DerivationPath, chainCfg)
	r.NoError(err)

	queryClient := mocks.SetupAuthClient(t)

	baseAcc := authTypes.ProtoBaseAccount()
	err = baseAcc.SetAddress(sdkTypes.AccAddress(wallet.AccAddress()))
	r.NoError(err)
	any, err := codectypes.NewAnyWithValue(baseAcc)
	r.NoError(err)
	queryResp := &authTypes.QueryAccountResponse{
		Account: any,
	}
	queryClient.EXPECT().Account(gomock.Any(), &authTypes.QueryAccountRequest{Address: wallet.AccAddress()}).Return(queryResp, nil).AnyTimes()

	wallet.Client.AuthClient = queryClient

	serviceClient := mocks.SetupServiceClient(t)

	// auto gas & fees
	simRes := txTypes.SimulateResponse{
		GasInfo: &sdkTypes.GasInfo{GasWanted: 0, GasUsed: 0},
		Result:  nil,
	}
	serviceClient.EXPECT().Simulate(gomock.Any(), gomock.Any()).Return(&simRes, nil).AnyTimes()
	wallet.Client.TxClient = serviceClient

	rpcClient := mocks.SetupRPCClient(t)
	re := coretypes.ResultStatus{
		NodeInfo: p2p.DefaultNodeInfo{Network: "jackaaaal"},
	}
	rpcClient.EXPECT().Status(gomock.Any()).Return(&re, nil).AnyTimes()
	rpcClient.EXPECT().BroadcastTxCommit(gomock.Any(), gomock.Any()).Return(nil, status.Error(codes.Aborted, "some error")).AnyTimes()
	wallet.Client.RPCClient = rpcClient

	chMsg := make(chan *Message)
	w := newWorker(0, wallet, 1, 10, 3, chMsg)

	wg := sync.WaitGroup{}
	wg.Add(1)

	msg := types.NewMsgPostProof(w.wallet.AccAddress(), []byte("hello"), "owner", 0, []byte("item"), []byte("list"), 0)
	m := Message{
		msg: msg,
		wg:  &wg,
	}

	w.batch = append(w.batch, &m)

	t.Log("sending msg")
	w.send()

	wg.Wait()
	r.ErrorIs(m.err, ErrReachedMaxRetry)
	close(chMsg)
}

func TestBatchFullSend(t *testing.T) {
	r := require.New(t)

	chainCfg := wtypes.ChainConfig{
		RPCAddr:       "http://localhost:26657",
		GRPCAddr:      "localhost:9090",
		GasPrice:      "0.02ujkl",
		GasAdjustment: 1.5,
		Bech32Prefix:  "jkl",
	}

	s := &config.Seed{
		SeedPhrase:     "forward service profit benefit punch catch fan chief jealous steel harvest column spell rude warm home melody hat broccoli pulse say garlic you firm",
		DerivationPath: "m/44'/118'/0'/0/0",
	}

	wallet, err := sequoiaWallet.CreateWallet(s.SeedPhrase, s.DerivationPath, chainCfg)
	r.NoError(err)

	queryClient := mocks.SetupAuthClient(t)

	baseAcc := authTypes.ProtoBaseAccount()
	err = baseAcc.SetAddress(sdkTypes.AccAddress(wallet.AccAddress()))
	r.NoError(err)
	any, err := codectypes.NewAnyWithValue(baseAcc)
	r.NoError(err)
	queryResp := &authTypes.QueryAccountResponse{
		Account: any,
	}
	queryClient.EXPECT().Account(gomock.Any(), &authTypes.QueryAccountRequest{Address: wallet.AccAddress()}).Return(queryResp, nil).AnyTimes()

	wallet.Client.AuthClient = queryClient

	serviceClient := mocks.SetupServiceClient(t)

	// auto gas & fees
	simRes := txTypes.SimulateResponse{
		GasInfo: &sdkTypes.GasInfo{GasWanted: 0, GasUsed: 0},
		Result:  nil,
	}
	serviceClient.EXPECT().Simulate(gomock.Any(), gomock.Any()).Return(&simRes, nil).AnyTimes()
	wallet.Client.TxClient = serviceClient

	rpcClient := mocks.SetupRPCClient(t)
	re := coretypes.ResultStatus{
		NodeInfo: p2p.DefaultNodeInfo{Network: "jackaaaal"},
	}
	rpcClient.EXPECT().Status(gomock.Any()).Return(&re, nil).AnyTimes()
	wallet.Client.RPCClient = rpcClient

	wg := sync.WaitGroup{}
	wg.Add(2)

	m0 := Message{
		msg: types.NewMsgPostProof("1", []byte(""), "", 0, []byte(""), []byte(""), 0),
		wg:  &wg,
	}
	m1 := Message{
		msg: types.NewMsgPostProof("2", []byte(""), "", 0, []byte(""), []byte(""), 0),
		wg:  &wg,
	}

	txData := wtypes.NewTransactionData(m0.msg, m1.msg).WithFeeAuto().WithGasAuto()

	txEncoder := tx.DefaultTxEncoder()
	builder, err := wallet.BuildTx(txData)
	r.NoError(err)

	expectedSignedTx, err := txEncoder(builder.GetTx())
	r.NoError(err)

	var actual ttypes.Tx

	rpcClient.
		EXPECT().
		BroadcastTxCommit(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, tx ttypes.Tx) (*sdkTypes.TxResponse, error) {
			actual = tx

			return nil, status.Error(codes.OK, "no error")
		}).
		AnyTimes()

	chMsg := make(chan *Message)
	w := newWorker(0, wallet, 3, 2, 2, chMsg)
	go w.start()

	chMsg <- &m0
	chMsg <- &m1

	wg.Wait()

	close(chMsg)
	r.EqualValues(expectedSignedTx, actual)
}
