package api

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"net/http"
	"sequoia/api/types"
	"sequoia/file_system"
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
				fmt.Println(err)
			}
			return
		}

		err = json.NewEncoder(w).Encode(dump)
		if err != nil {
			fmt.Println(err)
		}

	}
}
