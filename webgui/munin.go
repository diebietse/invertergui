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
	"bytes"
	"fmt"
	"net/http"

	"github.com/diebietse/invertergui/mk2driver"
)

type muninData struct {
	status       mk2driver.Mk2Info
	timesUpdated int
}

func (w *WebGui) ServeMuninHTTP(rw http.ResponseWriter, r *http.Request) {
	muninDat := <-w.muninRespChan
	if muninDat.timesUpdated == 0 {
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
		fmt.Printf("%v\n", err)
	}
}

func (w *WebGui) ServeMuninConfigHTTP(rw http.ResponseWriter, r *http.Request) {
	output := `multigraph in_batvolt
graph_title Battery Voltage
graph_vlabel Voltage (V)
graph_category inverter
graph_info Battery voltage

volt.info Voltage of battery
volt.label Voltage of battery (V)

multigraph in_batcharge
graph_title Battery Charge
graph_vlabel Charge (%)
graph_category inverter
graph_info Battery charge

charge.info Estimated charge of battery
charge.label Battery charge (%)

multigraph in_batcurrent
graph_title Battery Current
graph_vlabel Current (A)
graph_category inverter
graph_info Battery current

current.info Battery current
current.label Battery current (A)

multigraph in_batpower
graph_title Battery Power
graph_vlabel Power (W)
graph_category inverter
graph_info Battery power

power.info Battery power
power.label Battery power (W)

multigraph in_mainscurrent
graph_title Mains Current
graph_vlabel Current (A)
graph_category inverter
graph_info Mains current

currentin.info Input current
currentin.label Input current (A)
currentout.info Output current
currentout.label Output current (A)

multigraph in_mainsvoltage
graph_title Mains Voltage
graph_vlabel Voltage (V)
graph_category inverter
graph_info Mains voltage

voltagein.info Input voltage
voltagein.label Input voltage (V)
voltageout.info Output voltage
voltageout.label Output voltage (V)

multigraph in_mainspower
graph_title Mains Power
graph_vlabel Power (VA)
graph_category inverter
graph_info Mains power

powerin.info Input power
powerin.label Input power (VA)
powerout.info Output power
powerout.label Output power (VA)

multigraph in_mainsfreq
graph_title Mains frequency
graph_vlabel Frequency (Hz)
graph_category inverter
graph_info Mains frequency

freqin.info In frequency
freqin.label In frequency (Hz)
freqout.info Out frequency
freqout.label Out frequency (Hz)
`

	_, err := rw.Write([]byte(output))
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

//Munin only samples once every 5 minutes so averages have to be calculated for some values.
func calcMuninValues(muninDat *muninData, newStatus *mk2driver.Mk2Info) {
	muninDat.timesUpdated++
	muninVal := &muninDat.status
	muninVal.OutCurrent += newStatus.OutCurrent
	muninVal.InCurrent += newStatus.InCurrent
	muninVal.BatCurrent += newStatus.BatCurrent

	muninVal.OutVoltage += newStatus.OutVoltage
	muninVal.InVoltage += newStatus.InVoltage
	muninVal.BatVoltage += newStatus.BatVoltage

	muninVal.InFrequency = newStatus.InFrequency
	muninVal.OutFrequency = newStatus.OutFrequency

	muninVal.ChargeState = newStatus.ChargeState
}

func calcMuninAverages(muninDat *muninData) {
	muninVal := &muninDat.status
	muninVal.OutCurrent /= float64(muninDat.timesUpdated)
	muninVal.InCurrent /= float64(muninDat.timesUpdated)
	muninVal.BatCurrent /= float64(muninDat.timesUpdated)

	muninVal.OutVoltage /= float64(muninDat.timesUpdated)
	muninVal.InVoltage /= float64(muninDat.timesUpdated)
	muninVal.BatVoltage /= float64(muninDat.timesUpdated)
}

func zeroMuninValues(muninDat *muninData) {
	muninDat.timesUpdated = 0
	muninVal := &muninDat.status
	muninVal.OutCurrent = 0
	muninVal.InCurrent = 0
	muninVal.BatCurrent = 0

	muninVal.OutVoltage = 0
	muninVal.InVoltage = 0
	muninVal.BatVoltage = 0

	muninVal.InFrequency = 0
	muninVal.OutFrequency = 0

	muninVal.ChargeState = 0
}
