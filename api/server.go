package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/JackalLabs/sequoia/api/types"

	"github.com/rs/cors"

	"github.com/JackalLabs/sequoia/file_system"
	"github.com/JackalLabs/sequoia/recycle"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/JackalLabs/sequoia/proofs"
	"github.com/rs/zerolog/log"

	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/gorilla/mux"

	jsoniter "github.com/json-iterator/go"
)

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
	if a.srv == nil {
		return fmt.Errorf("no server available")
	}
	return a.srv.Close()
}

func (a *API) Serve(rd *recycle.RecycleDepot, f *file_system.FileSystem, p *proofs.Prover, wallet *wallet.Wallet, chunkSize int64) error {
	defer log.Info().Msg("API module stopped")
	r := mux.NewRouter()

	outline := types.NewOutline()

	outline.RegisterGetRoute(r, "/", IndexHandler(wallet.AccAddress()))

	outline.RegisterPostRoute(r, "/upload", PostFileHandler(f, p, wallet, chunkSize))
	outline.RegisterGetRoute(r, "/download/{merkle}", DownloadFileHandler(f))

	outline.RegisterGetRoute(r, "/list", ListFilesHandler(f))
	outline.RegisterGetRoute(r, "/api/client/list", ListFilesHandler(f))
	outline.RegisterGetRoute(r, "/api/data/fids", LegacyListFilesHandler(f))
	outline.RegisterGetRoute(r, "/api/client/space", SpaceHandler(wallet.Client, wallet.AccAddress()))

	outline.RegisterGetRoute(r, "/ipfs/peers", IPFSListPeers(f))
	outline.RegisterGetRoute(r, "/ipfs/hosts", IPFSListHosts(f))
	outline.RegisterGetRoute(r, "/ipfs/cids", IPFSListCids(f))
	outline.RegisterGetRoute(r, "/ipfs/cid_map", IPFSMapCids(f))
	outline.RegisterPostRoute(r, "/ipfs/make_folder", PostIPFSFolder(f))

	outline.RegisterGetRoute(r, "/dump", DumpDBHandler(f))

	outline.RegisterGetRoute(r, "/version", VersionHandler(wallet))
	outline.RegisterGetRoute(r, "/network", NetworkHandler(wallet))

	outline.RegisterGetRoute(r, "/recycle/salvage", RecycleSalvageHandler(rd))

	outline.RegisterGetRoute(r, "/api", outline.OutlineHandler())

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
