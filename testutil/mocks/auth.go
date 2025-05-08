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
