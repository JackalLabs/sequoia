package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/JackalLabs/sequoia/config"

	"github.com/JackalLabs/sequoia/api/types"

	"github.com/rs/cors"

	"github.com/JackalLabs/sequoia/file_system"
	storageTypes "github.com/jackalLabs/canine-chain/v4/x/storage/types"
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
	cfg  *config.APIConfig
}

// NewAPI creates a new API instance using the provided API configuration.
func NewAPI(cfg *config.APIConfig) *API {
	return &API{
		port: cfg.Port,
		cfg:  cfg,
	}
}

func (a *API) Close() error {
	if a.srv == nil {
		return fmt.Errorf("no server available")
	}
	return a.srv.Close()
}

func (a *API) Serve(f *file_system.FileSystem, p *proofs.Prover, wallet *wallet.Wallet, queryClient storageTypes.QueryClient, myIp string, chunkSize int64) {
	defer log.Info().Msg("API module stopped")
	r := mux.NewRouter()

	outline := types.NewOutline()

	outline.RegisterGetRoute(r, "/", IndexHandler(wallet.AccAddress()))

	outline.RegisterPostRoute(r, "/upload", PostFileHandler(f, p, wallet, queryClient, chunkSize))
	outline.RegisterPostRoute(r, "/v2/upload", PostFileHandlerV2(f, p, wallet, queryClient, chunkSize))
	outline.RegisterPostRoute(r, "/v2/status/{id}", CheckUploadStatus())
	outline.RegisterPostRoute(r, "/api/jobs", ListJobsHandler())
	outline.RegisterGetRoute(r, "/download/{merkle}", DownloadFileHandler(f))

	if a.cfg.OpenGateway {
		outline.RegisterGetRoute(r, "/get/{merkle}/{path:.*}", FindFileHandler(f, wallet, myIp))
		outline.RegisterGetRoute(r, "/get/{merkle}", FindFileHandler(f, wallet, myIp))
	}

	outline.RegisterGetRoute(r, "/list", ListFilesHandler(f))
	outline.RegisterGetRoute(r, "/api/client/list", ListFilesHandler(f))
	outline.RegisterGetRoute(r, "/api/data/fids", LegacyListFilesHandler(f))
	outline.RegisterGetRoute(r, "/api/client/space", SpaceHandler(wallet.Client, queryClient, wallet.AccAddress()))

	outline.RegisterGetRoute(r, "/ipfs/peers", IPFSListPeers(f))
	outline.RegisterGetRoute(r, "/ipfs/hosts", IPFSListHosts(f))
	outline.RegisterGetRoute(r, "/ipfs/cids", IPFSListCids(f))
	outline.RegisterGetRoute(r, "/ipfs/cid_map", IPFSMapCids(f))
	outline.RegisterPostRoute(r, "/ipfs/make_folder", PostIPFSFolder(f))

	// outline.RegisterGetRoute(r, "/dump", DumpDBHandler(f))

	outline.RegisterGetRoute(r, "/version", VersionHandler(wallet))
	outline.RegisterGetRoute(r, "/network", NetworkHandler(wallet))

	outline.RegisterGetRoute(r, "/api", outline.OutlineHandler())

	r.Handle("/metrics", promhttp.Handler())
	r.Use(loggingMiddleware)

	handler := cors.Default().Handler(r)

	a.srv = &http.Server{
		Handler: handler,
		Addr:    fmt.Sprintf("0.0.0.0:%d", a.port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 300 * time.Second, // Add this (5 minutes)
		ReadTimeout:  600 * time.Second, // Increase this (10 minutes)
		IdleTimeout:  120 * time.Second, // Add this (2 minutes)
	}

	log.Logger.Info().Msg(fmt.Sprintf("Sequoia API now listening on %s", a.srv.Addr))

	// Create a channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- a.srv.ListenAndServe()
	}()

	// Wait for server error
	err := <-serverErrors
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Warn().Err(err).Msg("server error")
		return
	}
}
