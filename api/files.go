package api

import (
	"fmt"
	"net/http"

	"github.com/JackalLabs/sequoia/api/types"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/rs/zerolog/log"
)

func ListFilesHandler(f *file_system.FileSystem) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		merkles, _, _, err := f.ListFiles()
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
			mm[i] = fmt.Sprintf("%s", merkle)
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

func LegacyListFilesHandler(f *file_system.FileSystem) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		merkles, owners, _, err := f.ListFiles()
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
				CID: fmt.Sprintf("%s", m),
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
