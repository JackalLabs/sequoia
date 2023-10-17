package api

import (
	"encoding/json"
	"fmt"
	"github.com/JackalLabs/sequoia/api/types"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/dgraph-io/badger/v4"
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
				fmt.Println(err)
			}
			return
		}

		err = json.NewEncoder(w).Encode(files)
		if err != nil {
			fmt.Println(err)
		}

	}
}
