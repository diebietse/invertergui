package mk2driver

import "time"

type Led int

const (
	LedMain Led = iota
	LedAbsorption
	LedBulk
	LedFloat
	LedInverter
	LedOverload
	LedLowBattery
	LedTemperature
)

var LedNames = map[Led]string{
	LedTemperature: "led_over_temp",
	LedLowBattery:  "led_bat_low",
	LedOverload:    "led_overload",
	LedInverter:    "led_inverter",
	LedFloat:       "led_float",
	LedBulk:        "led_bulk",
	LedAbsorption:  "led_absorb",
	LedMain:        "led_mains",
}

type LEDstate int

const (
	LedOff LEDstate = iota
	LedOn
	LedBlink
)

var StateNames = map[LEDstate]string{
	LedOff:   "off",
	LedOn:    "on",
	LedBlink: "blink",
}

type Mk2Info struct {
	// Will be marked as false if an error is detected.
	Valid bool

	Version uint32

	BatVoltage float64
	// Positive current == charging
	// Negative current == discharging
	BatCurrent float64

	// Input AC parameters
	InVoltage   float64
	InCurrent   float64
	InFrequency float64

	// Output AC parameters
	OutVoltage   float64
	OutCurrent   float64
	OutFrequency float64

	// Charge state 0.0 to 1.0
	ChargeState float64

	// List LEDs
	LEDs map[Led]LEDstate

	Errors []error

	Timestamp time.Time
}

type Mk2 interface {
	C() chan *Mk2Info
	Close()
}
