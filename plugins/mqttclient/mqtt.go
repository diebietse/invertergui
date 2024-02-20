package mqttclient

import (
	"encoding/json"
	"time"

	"github.com/diebietse/invertergui/mk2driver"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("ctx", "inverter-gui-mqtt")

const keepAlive = 5 * time.Second

// Config sets MQTT client configuration
type Config struct {
	Broker   string
	ClientID string
	Topic    string
	Username string
	Password string
}

// New creates an MQTT client that starts publishing MK2 data as it is received.
func New(mk2 mk2driver.Mk2, config Config) error {
	c := mqtt.NewClient(getOpts(config))
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	go func() {
		for e := range mk2.C() {
			if e.Valid {
				data, err := json.Marshal(e)
				if err != nil {
					log.Errorf("Could not parse data source: %v", err)
					continue
				}

				t := c.Publish(config.Topic, 0, false, data)
				t.Wait()
				if t.Error() != nil {
					log.Errorf("Could not publish data: %v", t.Error())
				}
			}
		}
	}()
	return nil
}

func getOpts(config Config) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.Broker)
	opts.SetClientID(config.ClientID)

	if config.Username != "" {
		opts.SetUsername(config.Username)
	}
	if config.Password != "" {
		opts.SetPassword(config.Password)
	}
	opts.SetKeepAlive(keepAlive)

	opts.SetOnConnectHandler(func(mqtt.Client) {
		log.Info("Client connected to broker")
	})
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Errorf("Client connection to broker lost: %v", err)

	})
	return opts
}
