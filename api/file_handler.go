package api

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/gorilla/mux"
	storageTypes "github.com/jackalLabs/canine-chain/v3/x/storage/types"
	"net/http"
	"sequoia/api/types"
	"sequoia/file_system"
	"sequoia/queue"
)

const MaxFileSize = 32 << 30

func PostFileHandler(db *badger.DB, q *queue.Queue, address string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseMultipartForm(MaxFileSize) // MAX file size lives here
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
		sender := req.Form.Get("sender")

		file, _, err := req.FormFile("file") // Retrieve the file from form data
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

		merkle, fid, cid, size, err := file_system.WriteFile(db, file, sender, address, "")
		if err != nil {
			v := types.ErrorResponse{
				Error: err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			err = json.NewEncoder(w).Encode(v)
			if err != nil {
				fmt.Println(err)
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
				fmt.Println(err)
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
				fmt.Println(err)
			}
		}

		resp := types.UploadResponse{
			CID: cid,
			FID: fid,
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func DownloadFileHandler(db *badger.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)

		fid := vars["fid"]

		file, err := file_system.GetFileDataByFID(db, fid)
		if err != nil {
			fmt.Println(err)
			v := types.ErrorResponse{
				Error: err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			err = json.NewEncoder(w).Encode(v)
			if err != nil {
				fmt.Println(err)
			}
		}

		_, err = w.Write(file)
		if err != nil {
			fmt.Println(err)
		}
	}
}
