package api

import (
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
