package api

import (
	"fmt"
	"github.com/JackalLabs/sequoia/queue"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/dgraph-io/badger/v4"
	"github.com/gorilla/mux"
	"net/http"
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

func (a *API) Serve(db *badger.DB, q *queue.Queue, wallet *wallet.Wallet) {
	r := mux.NewRouter()
	r.HandleFunc("/", IndexHandler(wallet.AccAddress()))
	r.HandleFunc("/upload", PostFileHandler(db, q, wallet.AccAddress()))
	r.HandleFunc("/download/{fid}", DownloadFileHandler(db))

	r.HandleFunc("/list", ListFilesHandler(db))
	r.HandleFunc("/dump", DumpDBHandler(db))

	r.HandleFunc("/version", VersionHandler(wallet))

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("0.0.0.0:%d", a.port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
