package monitoring

import (
	"errors"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	// Import for side effects
)

// PProf encapsulates the pprof HTTP server.
type PProf struct {
	srv     *http.Server
	listen  string
	started int32 // atomic boolean: 0 = not started, 1 = started
}

// NewPProf creates a new PProf instance.
// It initializes an http.Server for pprof endpoints.
func NewPProf(listenAddr string) *PProf {
	server := &http.Server{
		Addr:    listenAddr,
		Handler: nil, // This will use http.DefaultServeMux, where pprof registers.
	}

	return &PProf{
		srv:    server,
		listen: listenAddr,
	}
}

// Start begins listening for pprof requests in a new goroutine.
// It uses atomic operations to ensure the server only starts once, even if called multiple times.
func (p *PProf) Start() {
	// Use atomic compare-and-swap to ensure only one goroutine can start the server
	if !atomic.CompareAndSwapInt32(&p.started, 0, 1) {
		log.Debug().Msg("PProf server already started, ignoring duplicate Start call")
		return
	}

	go func() {
		log.Info().Str("addr", p.listen).Msg("Starting PProf server")
		err := p.srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("PProf server failed")
		} else if errors.Is(err, http.ErrServerClosed) {
			log.Info().Msg("PProf server stopped gracefully")
		}
	}()
}

// Stop shuts down the pprof server gracefully.
// It resets the started flag to allow the server to be restarted.
func (p *PProf) Stop() error {
	log.Info().Msg("Shutting down PProf server")
	err := p.srv.Close()
	// Reset the started flag to allow restart
	atomic.StoreInt32(&p.started, 0)
	return err
}
