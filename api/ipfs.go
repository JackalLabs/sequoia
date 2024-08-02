package api

import (
	"encoding/hex"
	"net/http"

	"github.com/JackalLabs/sequoia/api/types"
	"github.com/rs/zerolog/log"

	"github.com/JackalLabs/sequoia/file_system"
)

func IPFSListPeers(f *file_system.FileSystem) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		peers := f.ListPeers()
		f := types.PeersResponse{
			Peers: peers,
		}

		err := json.NewEncoder(w).Encode(f)
		if err != nil {
			log.Error().Err(err)
		}
	}
}

func IPFSListCids(f *file_system.FileSystem) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		cids, err := f.ListCids()
		if err != nil {
			handleErr(err, w, http.StatusInternalServerError)
			return
		}

		f := types.CidResponse{
			Cids: cids,
		}
		err = json.NewEncoder(w).Encode(f)
		if err != nil {
			log.Error().Err(err)
		}
	}
}

func IPFSMapCids(f *file_system.FileSystem) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		cids, err := f.MapCids()
		if err != nil {
			handleErr(err, w, http.StatusInternalServerError)
			return
		}

		m := make(map[string]string)

		for cid, merkle := range cids {
			merk := hex.EncodeToString(merkle)
			m[cid] = merk
		}

		f := types.CidMapResponse{
			CidMap: m,
		}
		err = json.NewEncoder(w).Encode(f)
		if err != nil {
			log.Error().Err(err)
		}
	}
}
