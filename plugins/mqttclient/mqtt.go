package mqttclient

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/diebietse/invertergui/mk2driver"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

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
					fmt.Printf("Data error: %v\n", err)
					continue
				}

				t := c.Publish(config.Topic, 0, false, data)
				t.Wait()
				if t.Error() != nil {
					fmt.Printf("Error: %v\n", t.Error())
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
		fmt.Print("Client connected to broker")
	})
	opts.SetConnectionLostHandler(func(cli mqtt.Client, err error) {
		fmt.Printf("Client connection to broker losted: %v", err)

	})
	return opts
}
