package api

import (
	"context"
	"net/http"

	"github.com/JackalLabs/sequoia/config"

	"github.com/JackalLabs/sequoia/api/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
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

func VersionHandler(wallet *wallet.Wallet) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		chainId, err := wallet.Client.GetChainID()
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

func NetworkHandler(wallet *wallet.Wallet) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		status, err := wallet.Client.RPCClient.Status(context.Background())
		if err != nil {
			w.WriteHeader(500)
			return
		}

		grpcStatus := wallet.Client.GRPCConn.GetState()

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
