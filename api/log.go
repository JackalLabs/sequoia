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

		length := 100
		var err error
		v := req.Header.Get("length")
		if v != "" {
			length, err = strconv.Atoi(v)
			if err != nil {
				v := types.ErrorResponse{
					Error: errors.New("failed to parse length").Error(),
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

		buf := make([]byte, length)
		read, err := readLogFile(logf, buf)
		buf = buf[:read]
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
		_, err = w.Write(buf)
		if err != nil {
			log.Error().Err(err)
		}
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
