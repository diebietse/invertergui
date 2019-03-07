package frontend

import (
	"github.com/rakyll/statik/fs"

	"log"
	"net/http"
)

func NewStatic() http.Handler {
	statikFs, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	return http.FileServer(statikFs)
}
