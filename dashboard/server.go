package dashboard

import (
	"fmt"
	"log"
	"net/http"
)

func RunThis() {
	fmt.Printf("running...")
	http.HandleFunc("/", serveFiles)
	log.Fatal(http.ListenAndServe(":3001", nil))
}

func serveFiles(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	p := "." + r.URL.Path
	if p == "./" {
		p = "./static/index.html"
	}
	http.ServeFile(w, r, p)
}
