package api

import (
	"net/http"

	"github.com/JackalLabs/sequoia/api/types"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/rs/zerolog/log"
)

func DumpDBHandler(f *file_system.FileSystem) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		dump, err := f.Dump()
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

		err = json.NewEncoder(w).Encode(dump)
		if err != nil {
			log.Error().Err(err)
		}
	}
}
