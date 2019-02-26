package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/hpdvanwyk/invertergui/webgui"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	addr := flag.String("addr", ":8080", "TCP address to listen on.")
	mk2 := NewMk2Mock()
	gui := webgui.NewWebGui(mk2)

	http.Handle("/ws", http.HandlerFunc(gui.ServeHub))
	http.Handle("/", gui)
	http.Handle("/js/controller.js", http.HandlerFunc(gui.ServeJS))
	http.Handle("/munin", http.HandlerFunc(gui.ServeMuninHTTP))
	http.Handle("/muninconfig", http.HandlerFunc(gui.ServeMuninConfigHTTP))
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
