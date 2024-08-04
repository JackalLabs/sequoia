package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/cors"

	"github.com/JackalLabs/sequoia/file_system"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/JackalLabs/sequoia/proofs"
	"github.com/rs/zerolog/log"

	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/gorilla/mux"
)
import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type API struct {
	port int64
	srv  *http.Server
}

func NewAPI(port int64) *API {
	return &API{
		port: port,
	}
}

func (a *API) Close() error {
	return a.srv.Close()
}

func (a *API) Serve(f *file_system.FileSystem, p *proofs.Prover, wallet *wallet.Wallet, chunkSize int64) error {
	defer log.Info().Msg("API module stopped")
	r := mux.NewRouter()
	r.HandleFunc("/", IndexHandler(wallet.AccAddress()))
	r.HandleFunc("/upload", PostFileHandler(f, p, wallet, chunkSize))
	r.HandleFunc("/download/{merkle}", DownloadFileHandler(f))

	r.HandleFunc("/list", ListFilesHandler(f))
	r.HandleFunc("/api/data/fids", LegacyListFilesHandler(f))

	r.HandleFunc("/ipfs/peers", IPFSListPeers(f))
	r.HandleFunc("/ipfs/hosts", IPFSListHosts(f))
	r.HandleFunc("/ipfs/cids", IPFSListCids(f))
	r.HandleFunc("/ipfs/cid_map", IPFSMapCids(f))
	r.HandleFunc("/ipfs/make_folder", PostIPFSFolder(f))

	r.HandleFunc("/dump", DumpDBHandler(f))

	r.HandleFunc("/version", VersionHandler(wallet))

	r.Handle("/metrics", promhttp.Handler())
	r.Use(loggingMiddleware)

	handler := cors.Default().Handler(r)

	a.srv = &http.Server{
		Handler: handler,
		Addr:    fmt.Sprintf("0.0.0.0:%d", a.port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Logger.Info().Msg(fmt.Sprintf("Sequoia API now listening on %s", a.srv.Addr))
	err := a.srv.ListenAndServe()
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}
