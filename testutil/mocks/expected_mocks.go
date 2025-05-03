package mocks

import (
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

type AuthQueryClient interface {
	authtypes.QueryClient
}

type ServiceClient interface {
	tx.ServiceClient
}

type RPCClient interface {
	rpcclient.Client
}
