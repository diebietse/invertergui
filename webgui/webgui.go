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

package webgui

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/diebietse/invertergui/mk2if"
	"github.com/diebietse/invertergui/websocket"
)

const (
	LedOff     = "dot-off"
	LedRed     = "dot-red"
	BlinkRed   = "blink-red"
	LedGreen   = "dot-green"
	BlinkGreen = "blink-green"
)

type WebGui struct {
	stopChan chan struct{}

	muninRespChan chan muninData
	poller        mk2if.Mk2If
	wg            sync.WaitGroup
	hub           *websocket.Hub

	pu *prometheusUpdater
}

func NewWebGui(source mk2if.Mk2If) *WebGui {
	w := new(WebGui)
	w.muninRespChan = make(chan muninData)
	w.stopChan = make(chan struct{})
	w.poller = source
	w.pu = newPrometheusUpdater()
	w.hub = websocket.NewHub()

	w.wg.Add(1)
	go w.dataPoll()
	return w
}

type templateInput struct {
	Error []error `json:"errors"`

	Date string `json:"date"`

	OutCurrent string `json:"output_current"`
	OutVoltage string `json:"output_voltage"`
	OutPower   string `json:"output_power"`

	InCurrent string `json:"input_current"`
	InVoltage string `json:"input_voltage"`
	InPower   string `json:"input_power"`

	InMinOut string

	BatVoltage string `json:"battery_voltage"`
	BatCurrent string `json:"battery_current"`
	BatPower   string `json:"battery_power"`
	BatCharge  string `json:"battery_charge"`

	InFreq  string `json:"input_frequency"`
	OutFreq string `json:"output_frequency"`

	LedMap map[string]string `json:"led_map"`
}

func (w *WebGui) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	http.ServeFile(rw, r, "./frontend/index.html")
}

func (w *WebGui) ServeJS(rw http.ResponseWriter, r *http.Request) {
	http.ServeFile(rw, r, "./frontend/js/controller.js")
}

func (w *WebGui) ServeHub(rw http.ResponseWriter, r *http.Request) {
	w.hub.ServeHTTP(rw, r)
}

func ledName(led mk2if.Led) string {
	name, ok := mk2if.LedNames[led]
	if !ok {
		return "Unknown led"
	}
	return name
}

func buildTemplateInput(status *mk2if.Mk2Info) *templateInput {
	outPower := status.OutVoltage * status.OutCurrent
	inPower := status.InCurrent * status.InVoltage

	tmpInput := &templateInput{
		Error:      status.Errors,
		Date:       status.Timestamp.Format(time.RFC1123Z),
		OutCurrent: fmt.Sprintf("%.2f", status.OutCurrent),
		OutVoltage: fmt.Sprintf("%.2f", status.OutVoltage),
		OutPower:   fmt.Sprintf("%.2f", outPower),
		InCurrent:  fmt.Sprintf("%.2f", status.InCurrent),
		InVoltage:  fmt.Sprintf("%.2f", status.InVoltage),
		InFreq:     fmt.Sprintf("%.2f", status.InFrequency),
		OutFreq:    fmt.Sprintf("%.2f", status.OutFrequency),
		InPower:    fmt.Sprintf("%.2f", inPower),

		InMinOut: fmt.Sprintf("%.2f", inPower-outPower),

		BatCurrent: fmt.Sprintf("%.2f", status.BatCurrent),
		BatVoltage: fmt.Sprintf("%.2f", status.BatVoltage),
		BatPower:   fmt.Sprintf("%.2f", status.BatVoltage*status.BatCurrent),
		BatCharge:  fmt.Sprintf("%.2f", status.ChargeState*100),

		LedMap: map[string]string{},
	}
	for k, v := range status.LEDs {
		if k == mk2if.LedOverload || k == mk2if.LedTemperature || k == mk2if.LedLowBattery {
			switch v {
			case mk2if.LedOn:
				tmpInput.LedMap[ledName(k)] = LedRed
			case mk2if.LedBlink:
				tmpInput.LedMap[ledName(k)] = BlinkRed
			default:
				tmpInput.LedMap[ledName(k)] = LedOff
			}
		} else {
			switch v {
			case mk2if.LedOn:
				tmpInput.LedMap[ledName(k)] = LedGreen
			case mk2if.LedBlink:
				tmpInput.LedMap[ledName(k)] = BlinkGreen
			default:
				tmpInput.LedMap[ledName(k)] = LedOff
			}
		}
	}
	return tmpInput
}

func (w *WebGui) Stop() {
	close(w.stopChan)
	w.wg.Wait()
}

// dataPoll waits for data from the w.poller channel. It will send its currently stored status
// to respChan if anything reads from it.
func (w *WebGui) dataPoll() {
	pollChan := w.poller.C()
	var muninValues muninData
	for {
		select {
		case s := <-pollChan:
			if s.Valid {
				calcMuninValues(&muninValues, s)
				w.pu.updatePrometheus(s)
				err := w.hub.Broadcast(buildTemplateInput(s))
				if err != nil {
					log.Printf("Could not send update to clients: %v", err)
				}
			}
		case w.muninRespChan <- muninValues:
			zeroMuninValues(&muninValues)
		case <-w.stopChan:
			w.wg.Done()
			return
		}
	}
}
