package mqttclient

import (
	"crypto/tls"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sabih15/TeleOpServer/internal/platform/config"
)

type Client struct {
	inner mqtt.Client
}

func New(cfg *config.Config) (*Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTT.Broker).
		SetClientID(cfg.MQTT.ClientID).
		SetUsername(cfg.MQTT.Username).
		SetPassword(cfg.MQTT.Password).
		SetTLSConfig(&tls.Config{}).
		SetConnectTimeout(10 * time.Second).
		SetAutoReconnect(true).
		SetCleanSession(false)

	c := mqtt.NewClient(opts)
	if tok := c.Connect(); tok.Wait() && tok.Error() != nil {
		return nil, fmt.Errorf("mqtt connect: %w", tok.Error())
	}

	return &Client{inner: c}, nil
}

// Subscribe registers a handler for the given topic (supports wildcards: + and #).
func (c *Client) Subscribe(topic string, handler func(topic string, payload []byte)) error {
	tok := c.inner.Subscribe(topic, 1, func(_ mqtt.Client, msg mqtt.Message) {
		handler(msg.Topic(), msg.Payload())
	})
	tok.Wait()
	return tok.Error()
}

// Disconnect cleanly closes the connection.
func (c *Client) Disconnect() {
	c.inner.Disconnect(500)
}
