package api

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/JackalLabs/sequoia/proofs"

	"github.com/desmos-labs/cosmos-go-wallet/wallet"

	"github.com/JackalLabs/sequoia/api/types"
	"github.com/JackalLabs/sequoia/file_system"
	"github.com/gorilla/mux"
	storageTypes "github.com/jackalLabs/canine-chain/v3/x/storage/types"
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
		log.Error().Err(err)
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
			handleErr(fmt.Errorf("failed to find file on chain: %w", err), w, http.StatusInternalServerError)
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

		size, err := fio.WriteFile(file, merkle, sender, startBlock, wl.AccAddress(), chunkSize)
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
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Error().Err(fmt.Errorf("can't encode json : %w", err))
		}

		_ = prover.PostProof(merkle, sender, startBlock, startBlock, time.Now())
	}
}

func ForwardDownload(merkle []byte, wallet *wallet.Wallet, w http.ResponseWriter) {
	cl := storageTypes.NewQueryClient(wallet.Client.GRPCConn)
	fReq := storageTypes.QueryFindFile{Merkle: merkle}

	fileRes, err := cl.FindFile(context.Background(), &fReq)
	if err != nil {
		v := types.ErrorResponse{
			Error: err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(v)
		return
	}

	ips := fileRes.ProviderIps
	var ipStrings []string
	err = json.Unmarshal([]byte(ips), ipStrings)
	if err != nil {
		v := types.ErrorResponse{
			Error: err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(v)
		return
	}

	for _, ipString := range ipStrings {
		u, err := url.Parse(ipString)
		if err != nil {
			continue
		}

		merkleString := hex.EncodeToString(merkle)

		hasU := u.JoinPath("has", merkleString)

		hcl := http.DefaultClient

		r, err := http.NewRequest("GET", hasU.String(), nil)
		if err != nil {
			continue
		}

		hRes, err := hcl.Do(r)
		if err != nil {
			continue
		}

		if hRes.StatusCode != 200 {
			continue
		}

		mU := u.JoinPath("download", merkleString)
		http.Redirect(w, r, mU.String(), http.StatusFound)
		return
	}

	v := types.ErrorResponse{
		Error: "cannot find file on network",
	}
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(v)
}

func DownloadFileHandler(f *file_system.FileSystem, wallet *wallet.Wallet) func(http.ResponseWriter, *http.Request) {
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
			return
		}

		file, err := f.GetFileDataByMerkle(merkle)
		if err != nil {
			ForwardDownload(merkle, wallet, w)
			return
		}

		_, _ = w.Write(file)
	}
}

func HasFileHandler(f *file_system.FileSystem) func(http.ResponseWriter, *http.Request) {
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
			return

		}

		found, err := f.HasFile(merkle)
		if err != nil {
			v := types.ErrorResponse{
				Error: err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(v)
			return
		}

		if found {
			w.WriteHeader(200)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}
}
