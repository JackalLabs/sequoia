package api

import (
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"github.com/gorilla/mux"
	"net/http"
	"sequoia/queue"
	"time"
)

type API struct {
	port int64
}

func NewAPI(port int64) *API {
	return &API{
		port: port,
	}
}

func (a *API) Serve(db *badger.DB, q *queue.Queue, address string) {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler())
	r.HandleFunc("/upload", PostFileHandler(db, q, address))

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("127.0.0.1:%d", a.port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
