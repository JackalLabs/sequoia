package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/JackalLabs/sequoia/utils"

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

var JobMap sync.Map

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

		proofTypeString := req.Form.Get("type")
		if len(proofTypeString) == 0 {
			proofTypeString = "0"
		}
		proofType, err := strconv.ParseInt(proofTypeString, 10, 64)
		if err != nil {
			handleErr(fmt.Errorf("cannot parse proof type: %w", err), w, http.StatusBadRequest)
			return
		}

		file, fh, err := req.FormFile("file") // Retrieve the file from form data
		if err != nil {
			handleErr(fmt.Errorf("cannot get file from form: %w", err), w, http.StatusBadRequest)
			return
		}
		//nolint:errcheck
		defer file.Close()
		//nolint:errcheck
		defer req.MultipartForm.RemoveAll()

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

		size, c, err := fio.WriteFile(file, merkle, sender, startBlock, chunkSize, proofType, utils.GetIPFSParams(&f))
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

func PostFileHandlerV2(fio *file_system.FileSystem, prover *proofs.Prover, wl *wallet.Wallet, chunkSize int64) func(http.ResponseWriter, *http.Request) {
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

		proofTypeString := req.Form.Get("type")
		if len(proofTypeString) == 0 {
			proofTypeString = "0"
		}
		proofType, err := strconv.ParseInt(proofTypeString, 10, 64)
		if err != nil {
			handleErr(fmt.Errorf("cannot parse proof type: %w", err), w, http.StatusBadRequest)
			return
		}

		file, fh, err := req.FormFile("file") // Retrieve the file from form data
		if err != nil {
			handleErr(fmt.Errorf("cannot get file from form: %w", err), w, http.StatusBadRequest)
			return
		}
		//nolint:errcheck
		defer file.Close()
		//nolint:errcheck
		defer req.MultipartForm.RemoveAll()

		readSize := fh.Size
		if readSize == 0 {
			handleErr(fmt.Errorf("file cannot be empty"), w, http.StatusBadRequest)
			return
		}

		s := sha256.New() // creating id
		_, _ = s.Write(merkle)
		_, _ = s.Write([]byte(sender))
		_, _ = s.Write([]byte(strconv.FormatInt(startBlock, 10)))
		jobId := hex.EncodeToString(s.Sum(nil))

		up := types.UploadResponseV2{
			Merkle:   merkle,
			Owner:    sender,
			Start:    startBlock,
			CID:      "",
			Progress: 10,
		}

		JobMap.Store(jobId, &up)

		resp := types.AcceptedUploadResponse{ // send accepted response
			JobID: jobId,
		}
		w.WriteHeader(202)
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Error().Err(fmt.Errorf("can't encode json : %w", err))
		}

		cl := storageTypes.NewQueryClient(wl.Client.GRPCConn)
		queryParams := storageTypes.QueryFile{
			Merkle: merkle,
			Owner:  sender,
			Start:  startBlock,
		}
		res, err := cl.File(context.Background(), &queryParams)
		if err != nil {
			log.Error().Err(fmt.Errorf("failed to find file on chain with merkle: %x, owner: %s, start: %d | %w", merkle, sender, startBlock, err))
			return
		}
		up.Progress = 30

		f := res.File

		if readSize != f.FileSize {
			log.Error().Err(fmt.Errorf("cannot accept form file that doesn't match the chain data %d != %d", readSize, f.FileSize))
			return
		}

		if hex.EncodeToString(f.Merkle) != merkleString {
			log.Error().Err(fmt.Errorf("cannot accept file that doesn't match the chain data %x != %x", f.Merkle, merkle))
			return
		}

		if len(f.Proofs) == int(f.MaxProofs) {
			if !f.ContainsProver(wl.AccAddress()) {
				log.Error().Err(fmt.Errorf("cannot accept file that I cannot claim"))
				return
			}
		}
		up.Progress = 40

		size, c, err := fio.WriteFileWithProgress(file, merkle, sender, startBlock, chunkSize, proofType, utils.GetIPFSParams(&f), &up)
		if err != nil {
			log.Error().Err(fmt.Errorf("failed to write file to disk: %w", err))
			return
		}
		up.CID = c

		if int64(size) != f.FileSize {
			log.Error().Err(fmt.Errorf("cannot accept file that doesn't match the chain data %d != %d", int64(size), f.FileSize))
			return
		}

		_ = prover.PostProof(merkle, sender, startBlock, startBlock, time.Now())

		go func() {
			time.Sleep(10 * time.Minute)
			log.Info().Str("jobId", jobId).Msg("Deleting job after 10-minute retention period")
			JobMap.Delete(jobId)
		}()
	}
}

func ListJobsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		// Set response headers
		w.Header().Set("Content-Type", "application/json")

		// Create a slice to store all jobs
		type JobInfo struct {
			ID  string                  `json:"id"`
			Job *types.UploadResponseV2 `json:"job"`
		}

		jobsList := make([]JobInfo, 0)

		// Iterate through all items in the JobMap
		JobMap.Range(func(key, value interface{}) bool {
			jobID := key.(string)
			jobData := value.(*types.UploadResponseV2)

			// Add to our list
			jobsList = append(jobsList, JobInfo{
				ID:  jobID,
				Job: jobData,
			})

			return true // continue iteration
		})

		// Create response object
		response := struct {
			Count int       `json:"count"`
			Jobs  []JobInfo `json:"jobs"`
		}{
			Count: len(jobsList),
			Jobs:  jobsList,
		}

		// Encode and return the response
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Error().Err(fmt.Errorf("can't encode json: %w", err))
			handleErr(err, w, http.StatusInternalServerError)
			return
		}
	}
}

func CheckUploadStatus() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		id, found := vars["id"]
		if !found {
			handleErr(errors.New("could not get id from url"), w, http.StatusBadRequest)
			return
		}

		u, ok := JobMap.Load(id)
		if !ok {
			handleErr(errors.New("could not job from id"), w, http.StatusNotFound)
			return
		}

		k := u.(*types.UploadResponseV2)
		err := json.NewEncoder(w).Encode(k)
		if err != nil {
			log.Error().Err(fmt.Errorf("can't encode json : %w", err))
		}
	}
}

func PostIPFSFolder(f *file_system.FileSystem) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		//nolint:errcheck
		defer req.Body.Close()

		var cidList map[string]string

		err = json.Unmarshal(body, &cidList)
		if err != nil {
			if err != nil {
				http.Error(w, "Error parsing request body", http.StatusInternalServerError)
				return
			}
		}

		childCIDs := make(map[string]cid.Cid)
		for key, s := range cidList {
			c, err := cid.Parse(s)
			if err != nil {
				http.Error(w, fmt.Sprintf("Could not parse %s", s), http.StatusInternalServerError)
				return
			}
			childCIDs[key] = c
		}

		root, err := f.CreateIPFSFolder(childCIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		f := types.CidFolderResponse{
			Cid:  root.Cid().String(),
			Data: root.RawData(),
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
			return
		}

		file, err := f.GetFileData(merkle)
		if err != nil {
			v := types.ErrorResponse{
				Error: err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(v)
			return
		}
		rs := bytes.NewReader(file)
		http.ServeContent(w, req, merkleString, time.Time{}, rs)
	}
}
