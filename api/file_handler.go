package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JackalLabs/sequoia/api/types"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/JackalLabs/sequoia/queue"
	"github.com/dgraph-io/badger/v4"
	"github.com/gorilla/mux"
	storageTypes "github.com/jackalLabs/canine-chain/v3/x/storage/types"
	"github.com/rs/zerolog/log"
)

const MaxFileSize = 32 << 30

func PostFileHandler(db *badger.DB, q *queue.Queue, address string, chunkSize int64) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseMultipartForm(MaxFileSize) // MAX file size lives here
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
		sender := req.Form.Get("sender")

		file, _, err := req.FormFile("file") // Retrieve the file from form data
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

		merkle, fid, cid, size, err := file_system.WriteFile(db, file, sender, address, "", chunkSize)
		if err != nil {
			v := types.ErrorResponse{
				Error: err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			err = json.NewEncoder(w).Encode(v)
			if err != nil {
				log.Error().Err(err)
			}
		}

		msg := storageTypes.NewMsgPostContract(
			address,
			sender,
			fmt.Sprintf("%d", size),
			fid,
			merkle,
		)

		if err := msg.ValidateBasic(); err != nil {
			v := types.ErrorResponse{
				Error: err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			err = json.NewEncoder(w).Encode(v)
			if err != nil {
				log.Error().Err(err)
			}
		}

		m, wg := q.Add(msg)

		wg.Wait() // wait for queue to process

		if m.Error() != nil {
			v := types.ErrorResponse{
				Error: m.Error().Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			err = json.NewEncoder(w).Encode(v)
			if err != nil {
				log.Error().Err(err)
			}
		}

		resp := types.UploadResponse{
			CID: cid,
			FID: fid,
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Error().Err(err)
		}
	}
}

func DownloadFileHandler(db *badger.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)

		fid := vars["fid"]

		file, err := file_system.GetFileDataByFID(db, fid)
		if err != nil {
			v := types.ErrorResponse{
				Error: err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(v)

		}

		_, _ = w.Write(file)
	}
}
