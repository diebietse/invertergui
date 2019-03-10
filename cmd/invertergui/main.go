/*
Copyright (c) 2015, 2017 Hendrik van Wyk
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

* Neither the name of invertergui nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/diebietse/invertergui/frontend"
	"github.com/diebietse/invertergui/mk2driver"
	"github.com/diebietse/invertergui/webgui"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tarm/serial"
)

func main() {
	addr := flag.String("addr", ":8080", "TCP address to listen on.")

	tcp := flag.Bool("tcp", false, "Use TCP instead of TTY")
	ip := flag.String("ip", "localhost:8139", "IP to connect when using tcp connection.")
	dev := flag.String("dev", "/dev/ttyUSB0", "TTY device to use.")
	flag.Parse()

	var p io.ReadWriteCloser
	var err error
	var tcpAddr *net.TCPAddr

	if *tcp {
		tcpAddr, err = net.ResolveTCPAddr("tcp", *ip)
		if err != nil {
			panic(err)
		}
		p, err = net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			panic(err)
		}
	} else {
		serialConfig := &serial.Config{Name: *dev, Baud: 2400}
		p, err = serial.OpenPort(serialConfig)
		if err != nil {
			panic(err)
		}
	}
	defer p.Close()
	mk2, err := mk2driver.NewMk2Connection(p)
	if err != nil {
		panic(err)
	}
	defer mk2.Close()

	gui := webgui.NewWebGui(mk2)

	http.Handle("/", frontend.NewStatic())
	http.Handle("/ws", http.HandlerFunc(gui.ServeHub))
	http.Handle("/munin", http.HandlerFunc(gui.ServeMuninHTTP))
	http.Handle("/muninconfig", http.HandlerFunc(gui.ServeMuninConfigHTTP))
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
