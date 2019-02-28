package frontend

import (
	"github.com/rakyll/statik/fs"

	"log"
	"net/http"
)

type static struct {
	http.FileSystem
}

func NewStatic() http.Handler {
	statikFs, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	return &static{
		FileSystem: statikFs,
	}
}

func (s *static) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.FileServer(s).ServeHTTP(w, r)
}
