/*
Copyright (c) 2017, Hendrik van Wyk
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

package prometheus

import (
	"github.com/diebietse/invertergui/mk2driver"
	"github.com/prometheus/client_golang/prometheus"
)

type Prometheus struct {
	mk2driver.Mk2
	batteryVoltage  prometheus.Gauge
	batteryCharge   prometheus.Gauge
	batteryCurrent  prometheus.Gauge
	batteryPower    prometheus.Gauge
	mainsCurrentIn  prometheus.Gauge
	mainsCurrentOut prometheus.Gauge
	mainsVoltageIn  prometheus.Gauge
	mainsVoltageOut prometheus.Gauge
	mainsPowerIn    prometheus.Gauge
	mainsPowerOut   prometheus.Gauge
	mainsFreqIn     prometheus.Gauge
	mainsFreqOut    prometheus.Gauge
}

func NewPrometheus(mk2 mk2driver.Mk2) {
	tmp := &Prometheus{
		Mk2: mk2,
		batteryVoltage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "battery_voltage_v",
			Help: "Voltage of the battery.",
		}),
		batteryCharge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "battery_charge_percentage",
			Help: "Remaining battery charge.",
		}),
		batteryCurrent: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "battery_current_a",
			Help: "Battery current.",
		}),
		batteryPower: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "battery_power_w",
			Help: "Battery power.",
		}),
		mainsCurrentIn: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mains_current_in_a",
			Help: "Mains current flowing into inverter",
		}),
		mainsCurrentOut: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mains_current_out_a",
			Help: "Mains current flowing out of inverter",
		}),
		mainsVoltageIn: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mains_voltage_in_v",
			Help: "Mains voltage at input of inverter",
		}),
		mainsVoltageOut: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mains_voltage_out_v",
			Help: "Mains voltage at output of inverter",
		}),
		mainsPowerIn: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mains_power_in_va",
			Help: "Mains power in",
		}),
		mainsPowerOut: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mains_power_out_va",
			Help: "Mains power out",
		}),
		mainsFreqIn: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mains_freq_in_hz",
			Help: "Mains frequency at inverter input",
		}),
		mainsFreqOut: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mains_freq_out_hz",
			Help: "Mains frequency at inverter output",
		}),
	}
	prometheus.MustRegister(
		tmp.batteryVoltage,
		tmp.batteryCharge,
		tmp.batteryCurrent,
		tmp.batteryPower,
		tmp.mainsCurrentIn,
		tmp.mainsCurrentOut,
		tmp.mainsVoltageIn,
		tmp.mainsVoltageOut,
		tmp.mainsPowerIn,
		tmp.mainsPowerOut,
		tmp.mainsFreqIn,
		tmp.mainsFreqOut,
	)

	go tmp.run()
}

func (p *Prometheus) run() {
	for e := range p.C() {
		if e.Valid {
			p.updatePrometheus(e)
		}
	}
}

func (p *Prometheus) updatePrometheus(newStatus *mk2driver.Mk2Info) {
	s := newStatus
	p.batteryVoltage.Set(s.BatVoltage)
	p.batteryCharge.Set(newStatus.ChargeState * 100)
	p.batteryCurrent.Set(s.BatCurrent)
	p.batteryPower.Set(s.BatVoltage * s.BatCurrent)
	p.mainsCurrentIn.Set(s.InCurrent)
	p.mainsCurrentOut.Set(s.OutCurrent)
	p.mainsVoltageIn.Set(s.InVoltage)
	p.mainsVoltageOut.Set(s.OutVoltage)
	p.mainsPowerIn.Set(s.InVoltage * s.InCurrent)
	p.mainsPowerOut.Set(s.OutVoltage * s.OutCurrent)
	p.mainsFreqIn.Set(s.InFrequency)
	p.mainsFreqOut.Set(s.OutFrequency)
}
