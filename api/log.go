package api

import (
	"errors"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Info().
			Str("method", req.Method).
			Str("url", req.URL.RequestURI()).
			Str("user_agent", req.UserAgent()).
			Str("remote_addr", req.RemoteAddr).
			Msg("incoming http request")
		next.ServeHTTP(w, req)
	})
}

func readLogFile(logFile *os.File, buf []byte) (read int, err error) {
	stat, err := logFile.Stat()
	if err != nil {
		return 0, err
	}
	size := stat.Size()
	offset := size - int64(len(buf))
	if offset < 0 {
		offset = 0
	}

	_, err = logFile.Seek(offset, io.SeekStart)
	if err != nil {
		return 0, err
	}

	read, err = logFile.Read(buf)
	if !errors.Is(err, io.EOF) {
		return read, err
	}
	return read, nil
}
