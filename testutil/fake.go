package testutil

import (
	"context"
	"github.com/jackalLabs/canine-chain/v4/x/storage/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryClient = (*FakeStorageQueryClient)(nil)

type FakeStorageQueryClient struct {
}

func NewFakeStorageQueryClient() types.QueryClient {
	return &FakeStorageQueryClient{}
}

func (f *FakeStorageQueryClient) Params(ctx context.Context, in *types.QueryParams, opts ...grpc.CallOption) (*types.QueryParamsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
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
	return nil, status.Error(codes.Unimplemented, "this is a fake storage query client")
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
