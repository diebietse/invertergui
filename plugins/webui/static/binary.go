package static

import (
	"github.com/rakyll/statik/fs"

	"log"
	"net/http"
)

// New exports the static part of the webgui that is served via statik
func New() http.Handler {
	statikFs, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	return http.FileServer(statikFs)
}
