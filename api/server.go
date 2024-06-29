package api

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/JackalLabs/sequoia/file_system"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/JackalLabs/sequoia/proofs"
	"github.com/rs/zerolog/log"

	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/gorilla/mux"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

//go:embed static
var assets embed.FS

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

func (a *API) Serve(f *file_system.FileSystem, p *proofs.Prover, wallet *wallet.Wallet, chunkSize int64) {
	r := mux.NewRouter()
	r.HandleFunc("/", IndexHandler(wallet.AccAddress()))
	r.HandleFunc("/upload", PostFileHandler(f, p, wallet, chunkSize))
	r.HandleFunc("/download/{merkle}", DownloadFileHandler(f))

	r.HandleFunc("/list", ListFilesHandler(f))
	r.HandleFunc("/api/data/fids", LegacyListFilesHandler(f))

	r.HandleFunc("/dump", DumpDBHandler(f))

	r.HandleFunc("/version", VersionHandler(wallet))

	r.Handle("/metrics", promhttp.Handler())

	r.HandleFunc("/withdraw", WithdrawHandler(wallet, p)).Methods("POST")

	html, _ := fs.Sub(assets, "static")
	fs := http.FileServer(http.FS(html))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	a.srv = &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("0.0.0.0:%d", a.port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Logger.Info().Msg(fmt.Sprintf("Sequoia API now listening on %s...", a.srv.Addr))
	err := a.srv.ListenAndServe()
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}
}
