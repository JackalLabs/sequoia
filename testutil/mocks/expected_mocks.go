package mocks

import (
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	storage "github.com/jackalLabs/canine-chain/v4/x/storage/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"google.golang.org/grpc"
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

type GRPCConn interface {
	grpc.ClientConnInterface
}

type StorageQueryClient interface {
	storage.QueryClient
}
