package mocks

import (
	"testing"

	"go.uber.org/mock/gomock"
)

func SetupAuthClient(t *testing.T) *MockAuthQueryClient {
	ctrl := gomock.NewController(t)
	return NewMockAuthQueryClient(ctrl)
}

func SetupServiceClient(t *testing.T) *MockServiceClient {
	ctrl := gomock.NewController(t)
	return NewMockServiceClient(ctrl)
}

func SetupRPCClient(t *testing.T) *MockRPCClient {
	ctrl := gomock.NewController(t)
	return NewMockRPCClient(ctrl)
}
