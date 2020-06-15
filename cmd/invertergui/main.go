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
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/diebietse/invertergui/mk2core"
	"github.com/diebietse/invertergui/mk2driver"
	"github.com/diebietse/invertergui/plugins/cli"
	"github.com/diebietse/invertergui/plugins/mqttclient"
	"github.com/diebietse/invertergui/plugins/munin"
	"github.com/diebietse/invertergui/plugins/prometheus"
	"github.com/diebietse/invertergui/plugins/webui"
	"github.com/diebietse/invertergui/plugins/webui/static"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

var log = logrus.WithField("ctx", "inverter-gui")

func main() {
	conf, err := parseConfig()
	if err != nil {
		os.Exit(1)
	}
	log.Info("Starting invertergui")

	mk2, err := getMk2Device(conf.Data.Source, conf.Data.Host, conf.Data.Device)
	if err != nil {
		log.Fatalf("Could not open data source: %v", err)
	}
	defer mk2.Close()

	core := mk2core.NewCore(mk2)

	if conf.Cli.Enabled {
		cli.NewCli(core.NewSubscription())
	}

	// Webgui
	gui := webui.NewWebGui(core.NewSubscription())
	http.Handle("/", static.New())
	http.Handle("/ws", http.HandlerFunc(gui.ServeHub))

	// Munin
	mu := munin.NewMunin(core.NewSubscription())
	http.Handle("/munin", http.HandlerFunc(mu.ServeMuninHTTP))
	http.Handle("/muninconfig", http.HandlerFunc(mu.ServeMuninConfigHTTP))

	// Prometheus
	prometheus.NewPrometheus(core.NewSubscription())
	http.Handle("/metrics", promhttp.Handler())

	// MQTT
	if conf.MQTT.Enabled {
		mqttConf := mqttclient.Config{
			Broker:   conf.MQTT.Broker,
			Topic:    conf.MQTT.Topic,
			ClientID: conf.MQTT.ClientID,
			Username: conf.MQTT.Username,
			Password: conf.MQTT.Password,
		}
		if err := mqttclient.New(core.NewSubscription(), mqttConf); err != nil {
			log.Fatalf("Could not setup MQTT client: %v", err)
		}
	}
	log.Infof("Invertergui web server starting on: %v", conf.Address)

	if err := http.ListenAndServe(conf.Address, nil); err != nil {
		log.Fatal(err)
	}
}

func getMk2Device(source, ip, dev string) (mk2driver.Mk2, error) {
	var p io.ReadWriteCloser
	var err error
	var tcpAddr *net.TCPAddr

	switch source {
	case "serial":
		serialConfig := &serial.Config{Name: dev, Baud: 2400}
		p, err = serial.OpenPort(serialConfig)
		if err != nil {
			return nil, err
		}
	case "tcp":
		tcpAddr, err = net.ResolveTCPAddr("tcp", ip)
		if err != nil {
			return nil, err
		}
		p, err = net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			return nil, err
		}
	case "mock":
		return mk2driver.NewMk2Mock(), nil
	default:
		return nil, fmt.Errorf("Invalid source selection: %v\nUse \"serial\", \"tcp\" or \"mock\"", source)
	}

	mk2, err := mk2driver.NewMk2Connection(p)
	if err != nil {
		return nil, err
	}

	return mk2, nil
}
