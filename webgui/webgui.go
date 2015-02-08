/*
Copyright (c) 2015, Hendrik van Wyk
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
	"github.com/hpdvanwyk/invertergui/datasource"
	"html/template"
	"net/http"
	"time"
)

const (
	Temperature = iota
	Low_battery
	Overload
	Inverter
	Float
	Bulk
	Absorption
	Mains
)

var leds = map[int]string{
	0: "Temperature",
	1: "Low battery",
	2: "Overload",
	3: "Inverter",
	4: "Float",
	5: "Bulk",
	6: "Absorption",
	7: "Mains",
}

type WebGui struct {
	source   datasource.DataSource
	reqChan  chan *statusError
	respChan chan statusError
	stopChan chan struct{}
	template *template.Template

	muninRespChan chan muninData
}

func NewWebGui(source datasource.DataSource, pollRate time.Duration, batteryCapacity float64) *WebGui {
	wg := new(WebGui)
	wg.source = source
	wg.reqChan = make(chan *statusError)
	wg.respChan = make(chan statusError)
	wg.muninRespChan = make(chan muninData)
	wg.stopChan = make(chan struct{})
	var err error
	wg.template, err = template.New("thegui").Parse(htmlTemplate)
	if err != nil {
		panic(err)
	}
	go wg.dataPoll(pollRate, batteryCapacity)
	return wg
}

//TemplateInput is exported to be used as an argument to the http template package.
type TemplateInput struct {
	Error error

	OutCurrent string
	OutVoltage string
	OutPower   string

	InCurrent string
	InVoltage string
	InPower   string

	InMinOut string

	BatVoltage string
	BatCurrent string
	BatPower   string
	BatCharge  string

	InFreq string

	Leds []string
}

func (w *WebGui) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	statusErr := <-w.respChan

	tmpInput := buildTemplateInput(&statusErr)

	err := w.template.Execute(rw, tmpInput)
	if err != nil {
		panic(err)
	}
}

func buildTemplateInput(statusErr *statusError) *TemplateInput {
	status := statusErr.status
	outPower := status.OutVoltage * status.OutCurrent
	inPower := status.InCurrent * status.InVoltage

	tmpInput := &TemplateInput{
		Error:      statusErr.err,
		OutCurrent: fmt.Sprintf("%.3f", status.OutCurrent),
		OutVoltage: fmt.Sprintf("%.3f", status.OutVoltage),
		OutPower:   fmt.Sprintf("%.3f", outPower),
		InCurrent:  fmt.Sprintf("%.3f", status.InCurrent),
		InVoltage:  fmt.Sprintf("%.3f", status.InVoltage),
		InFreq:     fmt.Sprintf("%.3f", status.InFreq),
		InPower:    fmt.Sprintf("%.3f", inPower),

		InMinOut: fmt.Sprintf("%.3f", inPower-outPower),

		BatCurrent: fmt.Sprintf("%.3f", status.BatCurrent),
		BatVoltage: fmt.Sprintf("%.3f", status.BatVoltage),
		BatPower:   fmt.Sprintf("%.3f", status.BatVoltage*status.BatCurrent),
		BatCharge:  fmt.Sprintf("%.3f", statusErr.chargeLevel),
	}
	for i := 7; i >= 0; i-- {
		if status.Leds[i] == 1 {
			tmpInput.Leds = append(tmpInput.Leds, leds[i])
		}
	}
	return tmpInput
}

func (w *WebGui) Stop() {
	close(w.stopChan)
}

type statusError struct {
	status      datasource.MultiplusStatus
	chargeLevel float64
	err         error
}

// dataPoll will issue a request for a new status every pollRate. It will send its currently stored status
// to respChan if anything reads from it.
func (w *WebGui) dataPoll(pollRate time.Duration, batteryCapacity float64) {
	ticker := time.NewTicker(pollRate)
	tracker := NewChargeTracker(batteryCapacity)
	var statusErr statusError
	var muninValues muninData
	go w.getStatus()
	gettingStatus := true
	for {
		select {
		case <-ticker.C:
			if gettingStatus == false {
				go w.getStatus()
				gettingStatus = true
			}
		case s := <-w.reqChan:
			if s.err != nil {
				statusErr.err = s.err
			} else {
				statusErr.status = s.status
				statusErr.err = nil
				tracker.Update(s.status.BatCurrent)
				if s.status.Leds[Float] == 1 {
					tracker.Reset()
				}
				statusErr.chargeLevel = tracker.CurrentLevel()
				calcMuninValues(&muninValues, &statusErr)
			}
			gettingStatus = false
		case w.respChan <- statusErr:
		case w.muninRespChan <- muninValues:
			zeroMuninValues(&muninValues)
		case <-w.stopChan:
			return
		}
	}
}

func (w *WebGui) getStatus() {
	statusErr := new(statusError)
	statusErr.err = w.source.GetData(&statusErr.status)
	w.reqChan <- statusErr
}
