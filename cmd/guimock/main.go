package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/diebietse/invertergui/frontend"
	"github.com/diebietse/invertergui/mk2if"
	"github.com/diebietse/invertergui/webgui"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	addr := flag.String("addr", ":8080", "TCP address to listen on.")
	mk2 := mk2if.NewMk2Mock()
	gui := webgui.NewWebGui(mk2)

	http.Handle("/", frontend.NewStatic())
	http.Handle("/ws", http.HandlerFunc(gui.ServeHub))

	http.Handle("/munin", http.HandlerFunc(gui.ServeMuninHTTP))
	http.Handle("/muninconfig", http.HandlerFunc(gui.ServeMuninConfigHTTP))
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
