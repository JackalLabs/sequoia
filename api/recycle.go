package api

import (
	"net/http"

	"github.com/JackalLabs/sequoia/api/types"
	"github.com/JackalLabs/sequoia/recycle"
	"github.com/rs/zerolog/log"
)

func RecycleSalvageHandler(rd *recycle.RecycleDepot) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		v := types.RecycleSalvageResponse{
			TotalJackalProviderFiles: rd.TotalJprovFiles,
			SalvagedFilesCount:       rd.SalvagedFilesCount,
			IsSalvageFinished:        rd.SalvagedFilesCount == rd.TotalJprovFiles,
		}

		err := json.NewEncoder(w).Encode(v)
		if err != nil {
			log.Error().Err(err)
			return
		}
	}

}
