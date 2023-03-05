package munin

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diebietse/invertergui/mk2driver"
)

func TestServer(t *testing.T) {

	mockMk2 := mk2driver.NewMk2Mock()
	muninServer := NewMunin(mockMk2)

	ts := httptest.NewServer(http.HandlerFunc(muninServer.ServeMuninHTTP))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
}
