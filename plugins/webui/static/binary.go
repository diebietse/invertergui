package static

import (
	"embed"
	"net/http"
)

//go:embed css js index.html favicon.ico
var content embed.FS

// New exports the static part of the webgui that is served via embed
func New() http.Handler {
	return http.FileServer(http.FS(content))
}
