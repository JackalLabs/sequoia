package api

import (
	"encoding/json"
	"github.com/JackalLabs/sequoia/api/types"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/dgraph-io/badger/v4"
	"github.com/rs/zerolog/log"
	"net/http"
)

func ListFilesHandler(db *badger.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		files, err := file_system.ListFiles(db)

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

		f := types.ListResponse{
			Files: files,
			Count: len(files),
		}

		err = json.NewEncoder(w).Encode(f)
		if err != nil {
			log.Error().Err(err)
		}

	}
}

func LegacyListFilesHandler(db *badger.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		files, err := file_system.ListFiles(db)

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

		cids := make([]types.LegacyAPIListValue, len(files))

		for i, file := range files {
			cids[i] = types.LegacyAPIListValue{
				CID: file,
				FID: "",
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
