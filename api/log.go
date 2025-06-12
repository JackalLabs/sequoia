package api

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

// loggingMiddleware returns an HTTP middleware that logs incoming request details at the debug level before passing control to the next handler.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Debug().
			Str("method", req.Method).
			Str("url", req.URL.RequestURI()).
			Str("user_agent", req.UserAgent()).
			Str("remote_addr", req.RemoteAddr).
			Msg("incoming http request")
		next.ServeHTTP(w, req)
	})
}
