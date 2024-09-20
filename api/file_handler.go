package api

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	cid "github.com/ipfs/go-cid"

	"github.com/JackalLabs/sequoia/proofs"

	"github.com/desmos-labs/cosmos-go-wallet/wallet"

	"github.com/JackalLabs/sequoia/api/types"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/gorilla/mux"
	storageTypes "github.com/jackalLabs/canine-chain/v4/x/storage/types"
	"github.com/rs/zerolog/log"
)

// const MaxFileSize = 32 << 30
const MaxFileSize = 0

func handleErr(err error, w http.ResponseWriter, code int) {
	v := types.ErrorResponse{
		Error: err.Error(),
	}
	w.WriteHeader(code)
	err = json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Error().Err(err).Msg("json encoder failed to write error response to http response writer")
	}
}

func PostFileHandler(fio *file_system.FileSystem, prover *proofs.Prover, wl *wallet.Wallet, chunkSize int64) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseMultipartForm(MaxFileSize) // MAX file size lives here
		if err != nil {
			handleErr(fmt.Errorf("cannot parse form %w", err), w, http.StatusBadRequest)
			return
		}
		sender := req.Form.Get("sender")
		merkleString := req.Form.Get("merkle")
		merkle, err := hex.DecodeString(merkleString)
		if err != nil {
			handleErr(fmt.Errorf("cannot parse merkle: %w", err), w, http.StatusBadRequest)
			return
		}

		startBlockString := req.Form.Get("start")
		startBlock, err := strconv.ParseInt(startBlockString, 10, 64)
		if err != nil {
			handleErr(fmt.Errorf("cannot parse start block: %w", err), w, http.StatusBadRequest)
			return
		}

		file, fh, err := req.FormFile("file") // Retrieve the file from form data
		if err != nil {
			handleErr(fmt.Errorf("cannot get file from form: %w", err), w, http.StatusBadRequest)
			return
		}

		readSize := fh.Size
		if readSize == 0 {
			handleErr(fmt.Errorf("file cannot be empty"), w, http.StatusBadRequest)
			return
		}

		cl := storageTypes.NewQueryClient(wl.Client.GRPCConn)
		queryParams := storageTypes.QueryFile{
			Merkle: merkle,
			Owner:  sender,
			Start:  startBlock,
		}
		res, err := cl.File(context.Background(), &queryParams)
		if err != nil {
			handleErr(fmt.Errorf("failed to find file on chain with merkle: %x, owner: %s, start: %d | %w", merkle, sender, startBlock, err), w, http.StatusInternalServerError)
			return
		}

		f := res.File

		if readSize != f.FileSize {
			handleErr(fmt.Errorf("cannot accept form file that doesn't match the chain data %d != %d", readSize, f.FileSize), w, http.StatusInternalServerError)
			return
		}

		if hex.EncodeToString(f.Merkle) != merkleString {
			handleErr(fmt.Errorf("cannot accept file that doesn't match the chain data %x != %x", f.Merkle, merkle), w, http.StatusInternalServerError)
			return
		}

		if len(f.Proofs) == int(f.MaxProofs) {
			if !f.ContainsProver(wl.AccAddress()) {
				handleErr(fmt.Errorf("cannot accept file that I cannot claim"), w, http.StatusInternalServerError)
				return
			}
		}

		size, c, err := fio.WriteFile(file, merkle, sender, startBlock, wl.AccAddress(), chunkSize)
		if err != nil {
			handleErr(fmt.Errorf("failed to write file to disk: %w", err), w, http.StatusInternalServerError)
			return
		}

		if int64(size) != f.FileSize {
			handleErr(fmt.Errorf("cannot accept file that doesn't match the chain data %d != %d", int64(size), f.FileSize), w, http.StatusInternalServerError)
			return
		}

		resp := types.UploadResponse{
			Merkle: merkle,
			Owner:  sender,
			Start:  startBlock,
			CID:    c,
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Error().Err(fmt.Errorf("can't encode json : %w", err))
		}

		_ = prover.PostProof(merkle, sender, startBlock, startBlock, time.Now())
	}
}

func PostIPFSFolder(f *file_system.FileSystem) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		cidList := strings.Split(string(body), ",")

		childCIDs := make([]cid.Cid, len(cidList))

		fmt.Println(cidList)

		for i, s := range cidList {
			c, err := cid.Parse(s)
			if err != nil {
				http.Error(w, fmt.Sprintf("Could not parse %s", s), http.StatusInternalServerError)
				return
			}
			childCIDs[i] = c
		}

		root, err := f.CreateIPFSFolder(childCIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		f := types.CidFolderResponse{
			Cid: root,
		}

		err = json.NewEncoder(w).Encode(f)
		if err != nil {
			log.Error().Err(err)
		}
	}
}

func DownloadFileHandler(f *file_system.FileSystem) func(http.ResponseWriter, *http.Request) {
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

		file, err := f.GetFileData(merkle)
		if err != nil {
			v := types.ErrorResponse{
				Error: err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(v)

		}
		rs := bytes.NewReader(file)
		http.ServeContent(w, req, merkleString, time.Now(), rs)
	}
}
