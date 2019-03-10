package mk2driver

import (
	"fmt"
	"time"
)

type mock struct {
	c chan *Mk2Info
}

func NewMk2Mock() Mk2 {
	tmp := &mock{
		c: make(chan *Mk2Info, 1),
	}
	go tmp.genMockValues()
	return tmp
}

func genBaseLeds(state LEDstate) map[Led]LEDstate {
	return map[Led]LEDstate{
		LedMain:        state,
		LedAbsorption:  state,
		LedBulk:        state,
		LedFloat:       state,
		LedInverter:    state,
		LedOverload:    state,
		LedLowBattery:  state,
		LedTemperature: state,
	}
}

func (m *mock) C() chan *Mk2Info {
	return m.c
}

func (m *mock) Close() {

}

func (m *mock) genMockValues() {
	mult := 1.0
	ledState := LedOff
	for {
		input := &Mk2Info{
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
