package api

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/JackalLabs/sequoia/api/types"
	"github.com/rs/zerolog/log"
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

// reads logs from the log file name
// passing "" (empty file name) will disable this api
func LogHandler(logFileName string) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if logFileName == "" {
			v := types.ErrorResponse{
				Error: errors.New("log api is disabled").Error(),
			}
			w.WriteHeader(http.StatusForbidden)
			err := json.NewEncoder(w).Encode(v)
			if err != nil {
				log.Error().Err(err)
			}
			return

		}
		bytes := 100
		var err error
		v := req.Header.Get("bytes")
		if v != "" {
			bytes, err = strconv.Atoi(v)
			if err != nil {
				v := types.ErrorResponse{
					Error: errors.New("failed to parse int").Error(),
				}
				w.WriteHeader(http.StatusBadRequest)
				err = json.NewEncoder(w).Encode(v)
				if err != nil {
					log.Error().Err(err)
				}
				return
			}
		}

		logf, err := os.Open(logFileName)
		defer logf.Close()
		if err != nil {
			v := types.ErrorResponse{
				Error: errors.New("failed to get logs").Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			err = json.NewEncoder(w).Encode(v)
			if err != nil {
				log.Error().Err(err)
			}
			return

		}

		buf := make([]byte, bytes)
		_, err = readLogFile(logf, buf)
		if err != nil {
			v := types.ErrorResponse{
				Error: errors.New("failed to get logs").Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			err = json.NewEncoder(w).Encode(v)
			if err != nil {
				log.Error().Err(err)
			}
			return
		}
		err = json.NewEncoder(w).Encode(string(buf))
		if err != nil {
			log.Error().Err(err)
		}
	})
}

func readLogFile(logFile *os.File, buf []byte) (read int, err error) {
	_, err = logFile.Seek(-int64(len(buf)), io.SeekEnd)
	if err != nil {
		return 0, err
	}

	read, err = logFile.Read(buf)
	if !errors.Is(err, io.EOF) {
		return read, err
	}
	return read, nil
}
