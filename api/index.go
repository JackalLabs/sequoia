package api

import (
	"context"
	"net/http"

	"github.com/JackalLabs/sequoia/config"
	"github.com/JackalLabs/sequoia/rpc"

	"github.com/JackalLabs/sequoia/api/types"
	"github.com/rs/zerolog/log"
)

func IndexHandler(address string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		v := types.IndexResponse{
			Status:  "online",
			Address: address,
		}

		err := json.NewEncoder(w).Encode(v)
		if err != nil {
			log.Error().Err(err)
			return
		}
	}
}

func VersionHandler(fc *rpc.FailoverClient) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		chainId, err := fc.Wallet().Client.GetChainID()
		if err != nil {
			w.WriteHeader(500)
			return
		}

		v := types.VersionResponse{
			Version: config.Version(),
			Commit:  config.Commit(),
			ChainID: chainId,
		}

		err = json.NewEncoder(w).Encode(v)
		if err != nil {
			log.Error().Err(err)
			return
		}
	}
}

func NetworkHandler(fc *rpc.FailoverClient) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		status, err := fc.RPCClient().Status(context.Background())
		if err != nil {
			w.WriteHeader(500)
			return
		}

		grpcStatus := fc.GRPCConn().GetState()

		v := types.NetworkResponse{
			GRPCStatus: grpcStatus.String(),
			RPCStatus:  status,
		}

		err = json.NewEncoder(w).Encode(v)
		if err != nil {
			log.Error().Err(err)
			return
		}
	}
}
