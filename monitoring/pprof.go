package monitoring

import (
	"errors"
	"net/http"
	_ "net/http/pprof"

	"github.com/rs/zerolog/log"
	// Import for side effects
)

// PProf encapsulates the pprof HTTP server.
type PProf struct {
	srv    *http.Server
	listen string
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
func (p *PProf) Start() {
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
func (p *PProf) Stop() error {
	log.Info().Msg("Shutting down PProf server")
	return p.srv.Close()
}
