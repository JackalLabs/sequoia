package testutil

import (
	"context"

	codec "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
	bytes "github.com/tendermint/tendermint/libs/bytes"
	log "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/p2p"
	rpc "github.com/tendermint/tendermint/rpc/client"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	types1 "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	zlog "github.com/rs/zerolog/log"

	"github.com/desmos-labs/cosmos-go-wallet/wallet"
)

var _ types.QueryClient = (*FakeStorageQueryClient)(nil)

type FakeStorageQueryClient struct {
}

func NewFakeStorageQueryClient() *FakeStorageQueryClient {
	return &FakeStorageQueryClient{}
}

func (f *FakeStorageQueryClient) Params(ctx context.Context, in *types.QueryParams, opts ...grpc.CallOption) (*types.QueryParamsResponse, error) {

	resp := types.QueryParamsResponse{
		Params: types.Params{
			AttestFormSize:         5,
			AttestMinToPass:        3,
			CheckWindow:            300,
			ChunkSize:              10240,
			CollateralPrice:        10000000000,
			DepositAccount:         "jkl1t35eusvx97953uk47r3z4ckwd2prkn3fay76r8",
			MaxContractAgeInBlocks: 100,
			MissesToBurn:           1,
			PolRatio:               40,
			PriceFeed:              "jklprice",
			PricePerTbPerMonth:     15,
			ProofWindow:            7200,
			ReferralCommission:     25,
		},
	}
	return &resp, nil
}

// Queries a File by merkle, owner, and start.
func (f *FakeStorageQueryClient) File(ctx context.Context, in *types.QueryFile, opts ...grpc.CallOption) (*types.QueryFileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a list of File items.
func (f *FakeStorageQueryClient) AllFiles(ctx context.Context, in *types.QueryAllFiles, opts ...grpc.CallOption) (*types.QueryAllFilesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a File by merkle, owner, and start.
func (f *FakeStorageQueryClient) FilesFromNote(ctx context.Context, in *types.QueryFilesFromNote, opts ...grpc.CallOption) (*types.QueryFilesFromNoteResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a list of open files by provider_address.
func (f *FakeStorageQueryClient) OpenFiles(ctx context.Context, in *types.QueryOpenFiles, opts ...grpc.CallOption) (*types.QueryAllFilesResponse, error) {
	zlog.Info().Str("query", "OpenFiles").Msg("fake storage query client")
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a list of files with only 1x redundancy by provider_address.
func (f *FakeStorageQueryClient) EndangeredFiles(ctx context.Context, in *types.QueryOpenFiles, opts ...grpc.CallOption) (*types.QueryAllFilesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a list of File items matching the merkle.
func (f *FakeStorageQueryClient) AllFilesByMerkle(ctx context.Context, in *types.QueryAllFilesByMerkle, opts ...grpc.CallOption) (*types.QueryAllFilesByMerkleResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a list of File items matching the owner.
func (f *FakeStorageQueryClient) AllFilesByOwner(ctx context.Context, in *types.QueryAllFilesByOwner, opts ...grpc.CallOption) (*types.QueryAllFilesByOwnerResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a Proof by provider_address, merkle, owner, and start.
func (f *FakeStorageQueryClient) Proof(ctx context.Context, in *types.QueryProof, opts ...grpc.CallOption) (*types.QueryProofResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a list of Proof items.
func (f *FakeStorageQueryClient) AllProofs(ctx context.Context, in *types.QueryAllProofs, opts ...grpc.CallOption) (*types.QueryAllProofsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a list of Proof items by provider_address.
func (f *FakeStorageQueryClient) ProofsByAddress(ctx context.Context, in *types.QueryProofsByAddress, opts ...grpc.CallOption) (*types.QueryProofsByAddressResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a Provider by address.
func (f *FakeStorageQueryClient) Provider(ctx context.Context, in *types.QueryProvider, opts ...grpc.CallOption) (*types.QueryProviderResponse, error) {
	resp := new(types.QueryProviderResponse)
	resp.Provider.Totalspace = "100000000"
	return resp, nil
}

// Queries a list of Provider items.
func (f *FakeStorageQueryClient) AllProviders(ctx context.Context, in *types.QueryAllProviders, opts ...grpc.CallOption) (*types.QueryAllProvidersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries an Attestation by prover, merkle, owner, and start.
func (f *FakeStorageQueryClient) Attestation(ctx context.Context, in *types.QueryAttestation, opts ...grpc.CallOption) (*types.QueryAttestationResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a list of Attestation items.
func (f *FakeStorageQueryClient) AllAttestations(ctx context.Context, in *types.QueryAllAttestations, opts ...grpc.CallOption) (*types.QueryAllAttestationsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a Report by prover, merkle, owner, and start.
func (f *FakeStorageQueryClient) Report(ctx context.Context, in *types.QueryReport, opts ...grpc.CallOption) (*types.QueryReportResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a list of Report items.
func (f *FakeStorageQueryClient) AllReports(ctx context.Context, in *types.QueryAllReports, opts ...grpc.CallOption) (*types.QueryAllReportsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries free space by address.
func (f *FakeStorageQueryClient) FreeSpace(ctx context.Context, in *types.QueryFreeSpace, opts ...grpc.CallOption) (*types.QueryFreeSpaceResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries how many files a provider has stored by address.
func (f *FakeStorageQueryClient) StoreCount(ctx context.Context, in *types.QueryStoreCount, opts ...grpc.CallOption) (*types.QueryStoreCountResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries where a file is located by merkle.
func (f *FakeStorageQueryClient) FindFile(ctx context.Context, in *types.QueryFindFile, opts ...grpc.CallOption) (*types.QueryFindFileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries free space by address.
func (f *FakeStorageQueryClient) GetClientFreeSpace(ctx context.Context, in *types.QueryClientFreeSpace, opts ...grpc.CallOption) (*types.QueryClientFreeSpaceResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a PayData by address.
func (f *FakeStorageQueryClient) GetPayData(ctx context.Context, in *types.QueryPayData, opts ...grpc.CallOption) (*types.QueryPayDataResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a StoragePaymentInfo by address.
func (f *FakeStorageQueryClient) StoragePaymentInfo(ctx context.Context, in *types.QueryStoragePaymentInfo, opts ...grpc.CallOption) (*types.QueryStoragePaymentInfoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a list of StoragePaymentInfo items.
func (f *FakeStorageQueryClient) AllStoragePaymentInfo(ctx context.Context, in *types.QueryAllStoragePaymentInfo, opts ...grpc.CallOption) (*types.QueryAllStoragePaymentInfoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries whether user can upload a file based on size.
func (f *FakeStorageQueryClient) FileUploadCheck(ctx context.Context, in *types.QueryFileUploadCheck, opts ...grpc.CallOption) (*types.QueryFileUploadCheckResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries current price of storage.
func (f *FakeStorageQueryClient) PriceCheck(ctx context.Context, in *types.QueryPriceCheck, opts ...grpc.CallOption) (*types.QueryPriceCheckResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries a list of ActiveProviders items.
func (f *FakeStorageQueryClient) ActiveProviders(ctx context.Context, in *types.QueryActiveProviders, opts ...grpc.CallOption) (*types.QueryActiveProvidersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries protocol storage space used and purchased.
func (f *FakeStorageQueryClient) StorageStats(ctx context.Context, in *types.QueryStorageStats, opts ...grpc.CallOption) (*types.QueryStorageStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries how much storage space is being used on the network at this time.
func (f *FakeStorageQueryClient) NetworkSize(ctx context.Context, in *types.QueryNetworkSize, opts ...grpc.CallOption) (*types.QueryNetworkSizeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries the amount of offered storage by active providers
func (f *FakeStorageQueryClient) AvailableSpace(ctx context.Context, in *types.QueryAvailableSpace, opts ...grpc.CallOption) (*types.QueryAvailableSpaceResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

// Queries protocol storage space used and purchased.
func (f *FakeStorageQueryClient) Gauges(ctx context.Context, in *types.QueryAllGauges, opts ...grpc.CallOption) (*types.QueryAllGaugesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
}

var _ auth.QueryClient = (*FakeAuthQueryClient)(nil)

type FakeAuthQueryClient struct {
	wallet *wallet.Wallet
}

func NewFakeAuthQueryClient(w *wallet.Wallet) *FakeAuthQueryClient {
	return &FakeAuthQueryClient{
		wallet: w,
	}
}

// Accounts returns all the existing accounts
//
// Since: cosmos-sdk 0.43
func (a *FakeAuthQueryClient) Accounts(ctx context.Context, in *auth.QueryAccountsRequest, opts ...grpc.CallOption) (*auth.QueryAccountsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake auth query client")
}

// Account returns account details based on address.
func (a *FakeAuthQueryClient) Account(ctx context.Context, in *auth.QueryAccountRequest, opts ...grpc.CallOption) (*auth.QueryAccountResponse, error) {
	baseAcc := auth.ProtoBaseAccount()
	err := baseAcc.SetAddress(sdk.AccAddress(a.wallet.AccAddress()))
	if err != nil {
		return nil, err
	}

	any, err := codec.NewAnyWithValue(baseAcc)
	if err != nil {
		return nil, err
	}

	resp := &auth.QueryAccountResponse{
		Account: any,
	}

	return resp, nil
}

// Params queries all parameters.
func (a *FakeAuthQueryClient) Params(ctx context.Context, in *auth.QueryParamsRequest, opts ...grpc.CallOption) (*auth.QueryParamsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake auth query client")
}

// ModuleAccountByName returns the module account info by module name
func (a *FakeAuthQueryClient) ModuleAccountByName(ctx context.Context, in *auth.QueryModuleAccountByNameRequest, opts ...grpc.CallOption) (*auth.QueryModuleAccountByNameResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake auth query client")
}

var _ tx.ServiceClient = (*FakeServiceClient)(nil)

type FakeServiceClient struct {
}

func NewFakeServiceClient() *FakeServiceClient {
	return &FakeServiceClient{}
}

// Simulate simulates executing a transaction for estimating gas usage.
// returns 0 gas wanted, 0 gas used
func (s *FakeServiceClient) Simulate(ctx context.Context, in *tx.SimulateRequest, opts ...grpc.CallOption) (*tx.SimulateResponse, error) {
	simRes := tx.SimulateResponse{
		GasInfo: &sdk.GasInfo{GasWanted: 0, GasUsed: 0},
		Result:  nil,
	}

	return &simRes, nil
}

// GetTx fetches a tx by hash.
func (s *FakeServiceClient) GetTx(ctx context.Context, in *tx.GetTxRequest, opts ...grpc.CallOption) (*tx.GetTxResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake service client")
}

// BroadcastTx broadcast transaction.
func (s *FakeServiceClient) BroadcastTx(ctx context.Context, in *tx.BroadcastTxRequest, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake service client")
}

// GetTxsEvent fetches txs by event.
func (s *FakeServiceClient) GetTxsEvent(ctx context.Context, in *tx.GetTxsEventRequest, opts ...grpc.CallOption) (*tx.GetTxsEventResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake service client")
}

// GetBlockWithTxs fetches a block with decoded txs.
//
// Since: cosmos-sdk 0.45.2
func (s *FakeServiceClient) GetBlockWithTxs(ctx context.Context, in *tx.GetBlockWithTxsRequest, opts ...grpc.CallOption) (*tx.GetBlockWithTxsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake service client")
}

var _ rpc.Client = (*FakeRPCClient)(nil)

type FakeRPCClient struct {
}

func NewFakeRPCClient() *FakeRPCClient {
	return &FakeRPCClient{}
}

// ABCIInfo mocks base method.
func (m *FakeRPCClient) ABCIInfo(arg0 context.Context) (*coretypes.ResultABCIInfo, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// ABCIQuery mocks base method.
func (m *FakeRPCClient) ABCIQuery(ctx context.Context, path string, data bytes.HexBytes) (*coretypes.ResultABCIQuery, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// ABCIQueryWithOptions mocks base method.
func (m *FakeRPCClient) ABCIQueryWithOptions(ctx context.Context, path string, data bytes.HexBytes, opts rpc.ABCIQueryOptions) (*coretypes.ResultABCIQuery, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// Block mocks base method.
func (m *FakeRPCClient) Block(ctx context.Context, height *int64) (*coretypes.ResultBlock, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// BlockByHash mocks base method.
func (m *FakeRPCClient) BlockByHash(ctx context.Context, hash []byte) (*coretypes.ResultBlock, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// BlockResults mocks base method.
func (m *FakeRPCClient) BlockResults(ctx context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// BlockSearch mocks base method.
func (m *FakeRPCClient) BlockSearch(ctx context.Context, query string, page, perPage *int, orderBy string) (*coretypes.ResultBlockSearch, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// BlockchainInfo mocks base method.
func (m *FakeRPCClient) BlockchainInfo(ctx context.Context, minHeight, maxHeight int64) (*coretypes.ResultBlockchainInfo, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// BroadcastEvidence mocks base method.
func (m *FakeRPCClient) BroadcastEvidence(arg0 context.Context, arg1 types1.Evidence) (*coretypes.ResultBroadcastEvidence, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// BroadcastTxAsync mocks base method.
func (m *FakeRPCClient) BroadcastTxAsync(arg0 context.Context, arg1 types1.Tx) (*coretypes.ResultBroadcastTx, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// BroadcastTxCommit mocks base method.
// returns nil, nil
func (m *FakeRPCClient) BroadcastTxCommit(arg0 context.Context, arg1 types1.Tx) (*coretypes.ResultBroadcastTxCommit, error) {

	res := new(coretypes.ResultBroadcastTxCommit)

	tx, err := GetTxDecoder()(arg1)
	if err != nil {
		panic(err)
	}

	msgs := tx.GetMsgs()
	for _, m := range msgs {
		if msg, ok := m.(*types.MsgAddClaimer); ok {
			zlog.Info().Type("msg", msg).Str("claim address", msg.ClaimAddress).Msg("msg sent to fake rpc")
		} else if msg, ok := m.(*types.MsgPostProof); ok {
			zlog.Info().
				Type("msg", msg).
				Str("creator", msg.Creator).
				Bytes("item", msg.Item).
				Hex("hash_list", msg.HashList).
				Hex("merkle", msg.Merkle).
				Str("owner", msg.Owner).
				Int64("start", msg.Start).
				Int64("to_prove", msg.ToProve).
				Msg("msg sent to fake rpc")
		}

	}

	return res, nil
}

// BroadcastTxSync mocks base method.
func (m *FakeRPCClient) BroadcastTxSync(arg0 context.Context, arg1 types1.Tx) (*coretypes.ResultBroadcastTx, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// CheckTx mocks base method.
func (m *FakeRPCClient) CheckTx(arg0 context.Context, arg1 types1.Tx) (*coretypes.ResultCheckTx, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// Commit mocks base method.
func (m *FakeRPCClient) Commit(ctx context.Context, height *int64) (*coretypes.ResultCommit, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// ConsensusParams mocks base method.
func (m *FakeRPCClient) ConsensusParams(ctx context.Context, height *int64) (*coretypes.ResultConsensusParams, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// ConsensusState mocks base method.
func (m *FakeRPCClient) ConsensusState(arg0 context.Context) (*coretypes.ResultConsensusState, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// DumpConsensusState mocks base method.
func (m *FakeRPCClient) DumpConsensusState(arg0 context.Context) (*coretypes.ResultDumpConsensusState, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// Genesis mocks base method.
func (m *FakeRPCClient) Genesis(arg0 context.Context) (*coretypes.ResultGenesis, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// GenesisChunked mocks base method.
func (m *FakeRPCClient) GenesisChunked(arg0 context.Context, arg1 uint) (*coretypes.ResultGenesisChunk, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// Health mocks base method.
func (m *FakeRPCClient) Health(arg0 context.Context) (*coretypes.ResultHealth, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// IsRunning mocks base method.
func (m *FakeRPCClient) IsRunning() bool {
	return false
}

// NetInfo mocks base method.
func (m *FakeRPCClient) NetInfo(arg0 context.Context) (*coretypes.ResultNetInfo, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// NumUnconfirmedTxs mocks base method.
func (m *FakeRPCClient) NumUnconfirmedTxs(arg0 context.Context) (*coretypes.ResultUnconfirmedTxs, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// OnReset mocks base method.
func (m *FakeRPCClient) OnReset() error {
	return status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// OnStart mocks base method.
func (m *FakeRPCClient) OnStart() error {
	return status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// OnStop mocks base method.
func (m *FakeRPCClient) OnStop() {
}

// Quit mocks base method.
func (m *FakeRPCClient) Quit() <-chan struct{} {
	return make(chan struct{})
}

// Reset mocks base method.
func (m *FakeRPCClient) Reset() error {
	return status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// SetLogger mocks base method.
func (m *FakeRPCClient) SetLogger(arg0 log.Logger) {
}

// Start mocks base method.
func (m *FakeRPCClient) Start() error {
	return status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// Status mocks base method.
// returns Network: "jackaaaal"
func (m *FakeRPCClient) Status(arg0 context.Context) (*coretypes.ResultStatus, error) {
	re := coretypes.ResultStatus{
		NodeInfo: p2p.DefaultNodeInfo{Network: "jackaaaal"},
	}
	return &re, nil
}

// Stop mocks base method.
func (m *FakeRPCClient) Stop() error {
	return status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// String mocks base method.
func (m *FakeRPCClient) String() string {
	return "fake RPCClient"
}

// Subscribe mocks base method.
func (m *FakeRPCClient) Subscribe(ctx context.Context, subscriber, query string, outCapacity ...int) (<-chan coretypes.ResultEvent, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// Tx mocks base method.
func (m *FakeRPCClient) Tx(ctx context.Context, hash []byte, prove bool) (*coretypes.ResultTx, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// TxSearch mocks base method.
func (m *FakeRPCClient) TxSearch(ctx context.Context, query string, prove bool, page, perPage *int, orderBy string) (*coretypes.ResultTxSearch, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// UnconfirmedTxs mocks base method.
func (m *FakeRPCClient) UnconfirmedTxs(ctx context.Context, limit *int) (*coretypes.ResultUnconfirmedTxs, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// Unsubscribe mocks base method.
func (m *FakeRPCClient) Unsubscribe(ctx context.Context, subscriber, query string) error {
	return status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// UnsubscribeAll mocks base method.
func (m *FakeRPCClient) UnsubscribeAll(ctx context.Context, subscriber string) error {
	return status.Error(codes.Unimplemented, "this is fake RPCClient")
}

// Validators mocks base method.
func (m *FakeRPCClient) Validators(ctx context.Context, height *int64, page, perPage *int) (*coretypes.ResultValidators, error) {
	return nil, status.Error(codes.Unimplemented, "this is fake RPCClient")
}
