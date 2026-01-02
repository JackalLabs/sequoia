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
	"github.com/JackalLabs/sequoia/rpc"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/JackalLabs/sequoia/proofs"
	"github.com/rs/zerolog/log"

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

func (a *API) Serve(f *file_system.FileSystem, p *proofs.Prover, fc *rpc.FailoverClient, chunkSize int64, myIp string) {
	defer log.Info().Msg("API module stopped")
	r := mux.NewRouter()

	outline := types.NewOutline()

	outline.RegisterGetRoute(r, "/", IndexHandler(fc.AccAddress()))

	outline.RegisterPostRoute(r, "/upload", PostFileHandler(f, p, fc, chunkSize))
	outline.RegisterPostRoute(r, "/v2/upload", PostFileHandlerV2(f, p, fc, chunkSize))
	outline.RegisterPostRoute(r, "/v2/status/{id}", CheckUploadStatus())
	outline.RegisterPostRoute(r, "/api/jobs", ListJobsHandler())
	outline.RegisterGetRoute(r, "/download/{merkle}", DownloadFileHandler(f))

	if a.cfg.OpenGateway {
		outline.RegisterGetRoute(r, "/get/{merkle}/{path:.*}", FindFileHandler(f, fc, myIp))
		outline.RegisterGetRoute(r, "/get/{merkle}", FindFileHandler(f, fc, myIp))
	}

	outline.RegisterGetRoute(r, "/list", ListFilesHandler(f))
	outline.RegisterGetRoute(r, "/api/client/list", ListFilesHandler(f))
	outline.RegisterGetRoute(r, "/api/data/fids", LegacyListFilesHandler(f))
	outline.RegisterGetRoute(r, "/api/client/space", SpaceHandler(fc))

	outline.RegisterGetRoute(r, "/ipfs/peers", IPFSListPeers(f))
	outline.RegisterGetRoute(r, "/ipfs/hosts", IPFSListHosts(f))
	outline.RegisterGetRoute(r, "/ipfs/cids", IPFSListCids(f))
	outline.RegisterGetRoute(r, "/ipfs/cid_map", IPFSMapCids(f))
	outline.RegisterPostRoute(r, "/ipfs/make_folder", PostIPFSFolder(f))

	// outline.RegisterGetRoute(r, "/dump", DumpDBHandler(f))

	outline.RegisterGetRoute(r, "/version", VersionHandler(fc))
	outline.RegisterGetRoute(r, "/network", NetworkHandler(fc))

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
	err := a.srv.ListenAndServe()
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Warn().Err(err)
			return
		}
	}
}
