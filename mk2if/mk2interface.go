package mk2if

const (
	LED_TEMPERATURE = 128
	LED_LOW_BATTERY = 64
	LED_OVERLOAD    = 32
	LED_INVERTER    = 16
	LED_FLOAT       = 8
	LED_BULK        = 4
	LED_ABSORPTION  = 2
	LED_MAIN        = 1
)

var LedNames = map[int]string{
	LED_TEMPERATURE: "Temperature",
	LED_LOW_BATTERY: "Low Battery",
	LED_OVERLOAD:    "Overload",
	LED_INVERTER:    "Inverter",
	LED_FLOAT:       "Float",
	LED_BULK:        "Bulk",
	LED_ABSORPTION:  "Absorbtion",
	LED_MAIN:        "Mains",
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

	// List only active LEDs
	LedListOn    []int
	LedListBlink []int

	Errors []error
}

type Mk2If interface {
	GetMk2Info() *Mk2Info
	C() chan *Mk2Info
	Close()
}
