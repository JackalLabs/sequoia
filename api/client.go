package api

import (
	"context"
	"net/http"

	"github.com/JackalLabs/sequoia/api/types"
	"github.com/JackalLabs/sequoia/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	storageTypes "github.com/jackalLabs/canine-chain/v5/x/storage/types"
	"github.com/rs/zerolog/log"
)

func SpaceHandler(fc *rpc.FailoverClient) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		queryClient := storageTypes.NewQueryClient(fc.GRPCConn())
		address := fc.AccAddress()

		params := &storageTypes.QueryProvider{
			Address: address,
		}
		res, err := queryClient.Provider(context.Background(), params)
		if err != nil {
			v := types.ErrorResponse{
				Error: err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			err = json.NewEncoder(w).Encode(v)
			if err != nil {
				log.Error().Err(err)
			}
			return
		}

		totalSpace := res.Provider.Totalspace

		fsparams := &storageTypes.QueryFreeSpace{
			Address: address,
		}
		fsres, err := queryClient.FreeSpace(context.Background(), fsparams)
		if err != nil {
			v := types.ErrorResponse{
				Error: err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			err = json.NewEncoder(w).Encode(v)
			if err != nil {
				log.Error().Err(err)
			}
			return
		}

		freeSpace := fsres.Space

		ttint, ok := sdk.NewIntFromString(totalSpace)
		if !ok {
			return
		}

		v := types.SpaceResponse{
			Total: ttint.Int64(),
			Free:  freeSpace,
			Used:  ttint.Int64() - freeSpace,
		}

		err = json.NewEncoder(w).Encode(v)
		if err != nil {
			log.Error().Err(err)
		}
	}
}
