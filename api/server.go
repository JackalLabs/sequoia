package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/JackalLabs/sequoia/proofs"
	"github.com/rs/zerolog/log"

	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/dgraph-io/badger/v4"
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

func (a *API) Serve(db *badger.DB, p *proofs.Prover, wallet *wallet.Wallet, chunkSize int64) {
	r := mux.NewRouter()
	r.HandleFunc("/", IndexHandler(wallet.AccAddress()))
	r.HandleFunc("/upload", PostFileHandler(db, p, wallet, chunkSize))
	r.HandleFunc("/download/{fid}", DownloadFileHandler(db))

	r.HandleFunc("/list", ListFilesHandler(db))
	r.HandleFunc("/api/data/fids", LegacyListFilesHandler(db))

	r.HandleFunc("/dump", DumpDBHandler(db))

	r.HandleFunc("/version", VersionHandler(wallet))

	r.Handle("/metrics", promhttp.Handler())

	a.srv = &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("0.0.0.0:%d", a.port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Logger.Info().Msg(fmt.Sprintf("Sequoia API now listening on %s", a.srv.Addr))
	err := a.srv.ListenAndServe()
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}
}
