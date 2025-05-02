package mocks

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"go.uber.org/mock/gomock"
)

type AuthQueryClient interface {
	authtypes.QueryClient
}

func SetupAuthClient(t *testing.T) *MockAuthQueryClient {
	ctrl := gomock.NewController(t)
	return NewMockAuthQueryClient(ctrl)
}
