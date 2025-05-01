// Only used to generate mocks
package mocks

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type AuthQueryClient interface {
	authtypes.QueryClient
}
