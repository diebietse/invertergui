package cli

import (
	"fmt"
	"log"

	"github.com/diebietse/invertergui/mk2driver"
)

type Cli struct {
	mk2driver.Mk2
}

func NewCli(mk2 mk2driver.Mk2) {
	newCli := &Cli{
		Mk2: mk2,
	}
	go newCli.run()
}

func (c *Cli) run() {
	for e := range c.C() {
		if e.Valid {
			printInfo(e)
		}
	}
}

func printInfo(info *mk2driver.Mk2Info) {
	out := fmt.Sprintf("Version: %v\n", info.Version)
	out += fmt.Sprintf("Bat Volt: %.2fV Bat Cur: %.2fA \n", info.BatVoltage, info.BatCurrent)
	out += fmt.Sprintf("In Volt: %.2fV In Cur: %.2fA In Freq %.2fHz\n", info.InVoltage, info.InCurrent, info.InFrequency)
	out += fmt.Sprintf("Out Volt: %.2fV Out Cur: %.2fA Out Freq %.2fHz\n", info.OutVoltage, info.OutCurrent, info.OutFrequency)
	out += fmt.Sprintf("In Power %.2fW Out Power %.2fW\n", info.InVoltage*info.InCurrent, info.OutVoltage*info.OutCurrent)
	out += fmt.Sprintf("Charge State: %.2f%%\n", info.ChargeState*100)
	out += "LEDs state:"
	for k, v := range info.LEDs {
		out += fmt.Sprintf(" %s %s", mk2driver.LedNames[k], mk2driver.StateNames[v])
	}

	out += "\nErrors:"
	for _, v := range info.Errors {
		out += " " + v.Error()
	}
	out += "\n"
	log.Printf("System Info: \n%v", out)
}
