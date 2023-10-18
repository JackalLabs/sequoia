package api

import (
	"encoding/json"
	"github.com/JackalLabs/sequoia/api/types"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/dgraph-io/badger/v4"
	"github.com/rs/zerolog/log"
	"net/http"
)

func DumpDBHandler(db *badger.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		dump, err := file_system.Dump(db)

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
