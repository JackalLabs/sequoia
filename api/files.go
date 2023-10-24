package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JackalLabs/sequoia/api/types"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/dgraph-io/badger/v4"
	"github.com/rs/zerolog/log"
)

func ListFilesHandler(db *badger.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		merkles, _, _, err := file_system.ListFiles(db)
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

		mm := make([]string, len(merkles))
		for i, merkle := range merkles {
			mm[i] = fmt.Sprintf("%x", merkle)
		}

		f := types.ListResponse{
			Files: mm,
			Count: len(mm),
		}

		err = json.NewEncoder(w).Encode(f)
		if err != nil {
			log.Error().Err(err)
		}
	}
}

func LegacyListFilesHandler(db *badger.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		merkles, owners, _, err := file_system.ListFiles(db)
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

		cids := make([]types.LegacyAPIListValue, len(merkles))

		for i, m := range merkles {
			cids[i] = types.LegacyAPIListValue{
				CID: fmt.Sprintf("%x", m),
				FID: owners[i],
			}
		}

		f := types.LegacyAPIListResponse{
			Data: cids,
		}

		err = json.NewEncoder(w).Encode(f)
		if err != nil {
			log.Error().Err(err)
		}
	}
}
