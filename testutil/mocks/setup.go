package mocks

import (
	"go.uber.org/mock/gomock"
)

func SetupAuthClient(t gomock.TestReporter) *MockAuthQueryClient {
	ctrl := gomock.NewController(t)
	return NewMockAuthQueryClient(ctrl)
}

func SetupServiceClient(t gomock.TestReporter) *MockServiceClient {
	ctrl := gomock.NewController(t)
	return NewMockServiceClient(ctrl)
}

func SetupRPCClient(t gomock.TestReporter) *MockRPCClient {
	ctrl := gomock.NewController(t)
	return NewMockRPCClient(ctrl)
}

func SetupGRPCConn(t gomock.TestReporter) *MockGRPCConn {
	ctrl := gomock.NewController(t)
	grpc := NewMockGRPCConn(ctrl)

	return grpc
}

func SetupStorageQueryClient(t gomock.TestReporter) StorageQueryClient {
	ctrl := gomock.NewController(t)

	return NewMockStorageQueryClient(ctrl)
}
