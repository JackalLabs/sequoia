package api

import "net/http"

func HomeHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("Hello world!"))
	}
}
