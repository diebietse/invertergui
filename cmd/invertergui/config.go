package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
)

type config struct {
	Address string `long:"address" env:"ADDRESS" default:":8080" description:"The IP/DNS and port of the machine that the application is running on."`
	Data    struct {
		Source string `long:"data.source" env:"DATA_SOURCE" default:"serial" description:"Set the source of data for the inverter gui. \"serial\", \"tcp\" or \"mock\""`
		Host   string `long:"data.host" env:"DATA_HOST" default:"localhost:8139" description:"Host to connect when source is set to tcp."`
		Device string `long:"data.device" env:"DATA_DEVICE" default:"/dev/ttyUSB0" description:"TTY device to use when source is set to serial."`
	}
	Cli struct {
		Enabled bool `long:"cli.enabled" env:"CLI_ENABLED" description:"Enable CLI output."`
	}
	MQTT struct {
		Enabled      bool   `long:"mqtt.enabled" env:"MQTT_ENABLED" description:"Enable MQTT publishing."`
		Broker       string `long:"mqtt.broker" env:"MQTT_BROKER" default:"tcp://localhost:1883" description:"Set the host port and scheme of the MQTT broker."`
		ClientID     string `long:"mqtt.client_id" env:"MQTT_CLIENT_ID" default:"interter-gui" description:"Set the client ID for the MQTT connection."`
		Topic        string `long:"mqtt.topic" env:"MQTT_TOPIC" default:"invertergui/updates" description:"Set the MQTT topic updates published to."`
		Username     string `long:"mqtt.username" env:"MQTT_USERNAME" default:"" description:"Set the MQTT username"`
		Password     string `long:"mqtt.password" env:"MQTT_PASSWORD" default:"" description:"Set the MQTT password"`
		PasswordFile string `long:"mqtt.password-file" env:"MQTT_PASSWORD_FILE" default:"" description:"Path to a file containing the MQTT password"`
	}
	Loglevel string `long:"loglevel" env:"LOGLEVEL" default:"info" description:"The log level to generate logs at. (\"panic\", \"fatal\", \"error\", \"warn\", \"info\", \"debug\", \"trace\")"`
}

func parseConfig() (*config, error) {
	conf := &config{}
	parser := flags.NewParser(conf, flags.Default)
	if _, err := parser.Parse(); err != nil {
		return nil, err
	}
	if err := resolvePasswordFile(conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func resolvePasswordFile(conf *config) error {
	if conf.MQTT.PasswordFile != "" && conf.MQTT.Password != "" {
		return fmt.Errorf("mqtt.password and mqtt.password-file are mutually exclusive")
	}
	if conf.MQTT.PasswordFile != "" {
		password, err := readPasswordFile(conf.MQTT.PasswordFile)
		if err != nil {
			return err
		}
		conf.MQTT.Password = password
	}
	return nil
}

func readPasswordFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("could not read MQTT password file: %w", err)
	}
	return strings.TrimRight(string(data), "\n\r"), nil
}
