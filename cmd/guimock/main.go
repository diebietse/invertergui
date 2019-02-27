package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/hpdvanwyk/invertergui/frontend"
	"github.com/hpdvanwyk/invertergui/webgui"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	addr := flag.String("addr", ":8080", "TCP address to listen on.")
	mk2 := NewMk2Mock()
	gui := webgui.NewWebGui(mk2)

	rootFs := http.FileServer(frontend.BinaryFileSystem("root"))
	http.Handle("/", rootFs)
	jsFs := http.FileServer(frontend.BinaryFileSystem("js"))
	http.Handle("/js/", http.StripPrefix("/js", jsFs))
	cssFs := http.FileServer(frontend.BinaryFileSystem("css"))
	http.Handle("/css/", http.StripPrefix("/css", cssFs))
	http.Handle("/ws", http.HandlerFunc(gui.ServeHub))

	http.Handle("/munin", http.HandlerFunc(gui.ServeMuninHTTP))
	http.Handle("/muninconfig", http.HandlerFunc(gui.ServeMuninConfigHTTP))
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
