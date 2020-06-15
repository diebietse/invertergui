package cli

import (
	"github.com/diebietse/invertergui/mk2driver"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("ctx", "inverter-gui-cli")

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
	log.Infof("Version: %v", info.Version)
	log.Infof("Bat Volt: %.2fV Bat Cur: %.2fA", info.BatVoltage, info.BatCurrent)
	log.Infof("In Volt: %.2fV In Cur: %.2fA In Freq %.2fHz", info.InVoltage, info.InCurrent, info.InFrequency)
	log.Infof("Out Volt: %.2fV Out Cur: %.2fA Out Freq %.2fHz", info.OutVoltage, info.OutCurrent, info.OutFrequency)
	log.Infof("In Power %.2fW Out Power %.2fW", info.InVoltage*info.InCurrent, info.OutVoltage*info.OutCurrent)
	log.Infof("Charge State: %.2f%%", info.ChargeState*100)
	log.Info("LEDs state:")
	for k, v := range info.LEDs {
		log.Infof(" %s %s", mk2driver.LedNames[k], mk2driver.StateNames[v])
	}

	if len(info.Errors) != 0 {
		log.Info("Errors:")
		for _, err := range info.Errors {
			log.Error(err)
		}
	}
}
