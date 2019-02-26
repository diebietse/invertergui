package main

import (
	"fmt"
	"time"

	"github.com/hpdvanwyk/invertergui/mk2if"
)

type mock struct {
	c chan *mk2if.Mk2Info
}

func NewMk2Mock() mk2if.Mk2If {
	tmp := &mock{
		c: make(chan *mk2if.Mk2Info, 1),
	}
	go tmp.genMockValues()
	return tmp
}

func genBaseLeds(state mk2if.LEDstate) map[mk2if.Led]mk2if.LEDstate {
	return map[mk2if.Led]mk2if.LEDstate{
		mk2if.LedMain:        state,
		mk2if.LedAbsorption:  state,
		mk2if.LedBulk:        state,
		mk2if.LedFloat:       state,
		mk2if.LedInverter:    state,
		mk2if.LedOverload:    state,
		mk2if.LedLowBattery:  state,
		mk2if.LedTemperature: state,
	}
}

func (m *mock) GetMk2Info() *mk2if.Mk2Info {
	return &mk2if.Mk2Info{
		OutCurrent:   2.0,
		InCurrent:    2.3,
		OutVoltage:   230.0,
		InVoltage:    230.1,
		BatVoltage:   25,
		BatCurrent:   -10,
		InFrequency:  50,
		OutFrequency: 50,
		ChargeState:  1,
		Errors:       nil,
		Timestamp:    time.Now(),
		LEDs:         genBaseLeds(mk2if.LedOff),
	}
}

func (m *mock) C() chan *mk2if.Mk2Info {
	return m.c
}

func (m *mock) Close() {

}

func (m *mock) genMockValues() {
	mult := 1.0
	ledState := mk2if.LedOff
	for {
		input := &mk2if.Mk2Info{
			OutCurrent:   2.0 * mult,
			InCurrent:    2.3 * mult,
			OutVoltage:   230.0 * mult,
			InVoltage:    230.1 * mult,
			BatVoltage:   25 * mult,
			BatCurrent:   -10 * mult,
			InFrequency:  50 * mult,
			OutFrequency: 50 * mult,
			ChargeState:  1 * mult,
			Errors:       nil,
			Timestamp:    time.Now(),
			Valid:        true,
			LEDs:         genBaseLeds(ledState),
		}

		ledState = (ledState + 1) % 3

		mult = mult - 0.1
		if mult < 0 {
			mult = 1.0
		}
		fmt.Printf("Sending\n")
		m.c <- input
		time.Sleep(1 * time.Second)
	}
}
