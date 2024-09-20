package types

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type APIOutline struct {
	Routes []RouteOutline `json:"routes"`
}

type RouteOutline struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

func NewOutline() *APIOutline {
	return &APIOutline{
		Routes: make([]RouteOutline, 0),
	}
}

func (o *APIOutline) RegisterRoute(router *mux.Router, method string, path string, f func(http.ResponseWriter, *http.Request)) {
	o.Routes = append(o.Routes, RouteOutline{
		Method: method,
		Path:   path,
	})
	router.HandleFunc(path, f)
}

func (o *APIOutline) RegisterGetRoute(router *mux.Router, path string, f func(http.ResponseWriter, *http.Request)) {
	o.RegisterRoute(router, "GET", path, f)
}

func (o *APIOutline) RegisterPostRoute(router *mux.Router, path string, f func(http.ResponseWriter, *http.Request)) {
	o.RegisterRoute(router, "POST", path, f)
}

func (o *APIOutline) OutlineHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		err := json.NewEncoder(w).Encode(o.Routes)
		if err != nil {
			log.Error().Err(err)
			return
		}
	}
}
