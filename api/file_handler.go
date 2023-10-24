package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/desmos-labs/cosmos-go-wallet/wallet"

	"github.com/JackalLabs/sequoia/api/types"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/JackalLabs/sequoia/queue"
	"github.com/dgraph-io/badger/v4"
	"github.com/gorilla/mux"
	storageTypes "github.com/jackalLabs/canine-chain/v3/x/storage/types"
	"github.com/rs/zerolog/log"
)

const MaxFileSize = 32 << 30

func handleErr(err error, w http.ResponseWriter) {
	v := types.ErrorResponse{
		Error: err.Error(),
	}
	w.WriteHeader(http.StatusInternalServerError)
	err = json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Error().Err(err)
	}
}

func PostFileHandler(db *badger.DB, q *queue.Queue, wl *wallet.Wallet, chunkSize int64) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseMultipartForm(MaxFileSize) // MAX file size lives here
		if err != nil {
			handleErr(err, w)
			return
		}
		sender := req.Form.Get("sender")
		merkleString := req.Form.Get("merkle")
		merkle, err := hex.DecodeString(merkleString)
		if err != nil {
			handleErr(err, w)
			return
		}

		startBlockString := req.Form.Get("start")
		startBlock, err := strconv.ParseInt(startBlockString, 10, 64)
		if err != nil {
			handleErr(err, w)
			return
		}

		file, _, err := req.FormFile("file") // Retrieve the file from form data
		if err != nil {
			handleErr(err, w)
			return
		}

		cl := storageTypes.NewQueryClient(wl.Client.GRPCConn)
		queryParams := storageTypes.QueryFileRequest{
			Merkle: merkle,
			Owner:  sender,
			Start:  startBlock,
		}
		res, err := cl.File(context.Background(), &queryParams)
		if err != nil {
			handleErr(err, w)
			return
		}

		f := res.File

		if hex.EncodeToString(f.Merkle) != merkleString {
			handleErr(fmt.Errorf("cannot accept file that doesn't match the chain data %x != %x", f.Merkle, merkle), w)
			return
		}

		size, err := file_system.WriteFile(db, file, merkle, sender, startBlock, wl.AccAddress(), chunkSize)
		if err != nil {
			handleErr(err, w)
			return
		}

		if int64(size) != f.FileSize {
			handleErr(fmt.Errorf("cannot accept file that doesn't match the chain data %d != %d", int64(size), f.FileSize), w)
			return
		}

		resp := types.UploadResponse{
			Merkle: merkle,
			Owner:  sender,
			Start:  startBlock,
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

		merkleString := vars["merkle"]
		merkle, err := hex.DecodeString(merkleString)
		if err != nil {
			v := types.ErrorResponse{
				Error: err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(v)

		}

		file, err := file_system.GetFileDataByMerkle(db, merkle)
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
