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

package munin

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/diebietse/invertergui/mk2driver"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("ctx", "inverter-gui-munin")

type Munin struct {
	mk2driver.Mk2
	muninResponse chan muninData
}

type muninData struct {
	status       mk2driver.Mk2Info
	timesUpdated int
}

func NewMunin(mk2 mk2driver.Mk2) *Munin {
	m := &Munin{
		Mk2:           mk2,
		muninResponse: make(chan muninData),
	}

	go m.run()

	return m
}

func (m *Munin) ServeMuninHTTP(rw http.ResponseWriter, r *http.Request) {
	muninDat := <-m.muninResponse
	if muninDat.timesUpdated == 0 {
		log.Error("No data returned")
		rw.WriteHeader(500)
		_, _ = rw.Write([]byte("No data to return.\n"))
		return
	}
	calcMuninAverages(&muninDat)

	status := muninDat.status
	tmpInput := buildTemplateInput(&status)
	outputBuf := &bytes.Buffer{}
	fmt.Fprintf(outputBuf, "multigraph in_batvolt\n")
	fmt.Fprintf(outputBuf, "volt.value %s\n", tmpInput.BatVoltage)
	fmt.Fprintf(outputBuf, "multigraph in_batcharge\n")
	fmt.Fprintf(outputBuf, "charge.value %s\n", tmpInput.BatCharge)
	fmt.Fprintf(outputBuf, "multigraph in_batcurrent\n")
	fmt.Fprintf(outputBuf, "current.value %s\n", tmpInput.BatCurrent)
	fmt.Fprintf(outputBuf, "multigraph in_batpower\n")
	fmt.Fprintf(outputBuf, "power.value %s\n", tmpInput.BatPower)
	fmt.Fprintf(outputBuf, "multigraph in_mainscurrent\n")
	fmt.Fprintf(outputBuf, "currentin.value %s\n", tmpInput.InCurrent)
	fmt.Fprintf(outputBuf, "currentout.value %s\n", tmpInput.OutCurrent)
	fmt.Fprintf(outputBuf, "multigraph in_mainsvoltage\n")
	fmt.Fprintf(outputBuf, "voltagein.value %s\n", tmpInput.InVoltage)
	fmt.Fprintf(outputBuf, "voltageout.value %s\n", tmpInput.OutVoltage)
	fmt.Fprintf(outputBuf, "multigraph in_mainspower\n")
	fmt.Fprintf(outputBuf, "powerin.value %s\n", tmpInput.InPower)
	fmt.Fprintf(outputBuf, "powerout.value %s\n", tmpInput.OutPower)
	fmt.Fprintf(outputBuf, "multigraph in_mainsfreq\n")
	fmt.Fprintf(outputBuf, "freqin.value %s\n", tmpInput.InFreq)
	fmt.Fprintf(outputBuf, "freqout.value %s\n", tmpInput.OutFreq)

	_, err := rw.Write(outputBuf.Bytes())
	if err != nil {
		log.Errorf("Could not write data response: %v", err)
	}
}

func (m *Munin) ServeMuninConfigHTTP(rw http.ResponseWriter, r *http.Request) {
	output := muninConfig
	_, err := rw.Write([]byte(output))
	if err != nil {
		log.Errorf("Could not write config response: %v", err)
	}
}

func (m *Munin) run() {
	muninValues := &muninData{
		status: mk2driver.Mk2Info{},
	}
	for {
		select {
		case e := <-m.C():
			if e.Valid {
				calcMuninValues(muninValues, e)
			}
		case m.muninResponse <- *muninValues:
			zeroMuninValues(muninValues)
		}
	}
}

// Munin only samples once every 5 minutes so averages have to be calculated for some values.
func calcMuninValues(m *muninData, newStatus *mk2driver.Mk2Info) {
	m.timesUpdated++
	m.status.OutCurrent += newStatus.OutCurrent
	m.status.InCurrent += newStatus.InCurrent
	m.status.BatCurrent += newStatus.BatCurrent

	m.status.OutVoltage += newStatus.OutVoltage
	m.status.InVoltage += newStatus.InVoltage
	m.status.BatVoltage += newStatus.BatVoltage

	m.status.InFrequency = newStatus.InFrequency
	m.status.OutFrequency = newStatus.OutFrequency

	m.status.ChargeState = newStatus.ChargeState
}

func calcMuninAverages(m *muninData) {
	m.status.OutCurrent /= float64(m.timesUpdated)
	m.status.InCurrent /= float64(m.timesUpdated)
	m.status.BatCurrent /= float64(m.timesUpdated)

	m.status.OutVoltage /= float64(m.timesUpdated)
	m.status.InVoltage /= float64(m.timesUpdated)
	m.status.BatVoltage /= float64(m.timesUpdated)
}

func zeroMuninValues(m *muninData) {
	m.timesUpdated = 0
	m.status.OutCurrent = 0
	m.status.InCurrent = 0
	m.status.BatCurrent = 0

	m.status.OutVoltage = 0
	m.status.InVoltage = 0
	m.status.BatVoltage = 0

	m.status.InFrequency = 0
	m.status.OutFrequency = 0

	m.status.ChargeState = 0
}

type templateInput struct {
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
}

func buildTemplateInput(status *mk2driver.Mk2Info) *templateInput {
	outPower := status.OutVoltage * status.OutCurrent
	inPower := status.InCurrent * status.InVoltage

	newInput := &templateInput{
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
	}
	return newInput
}
