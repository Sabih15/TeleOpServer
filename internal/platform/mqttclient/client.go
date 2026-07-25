package mqttclient

import (
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sabih15/TeleOpServer/internal/platform/config"
)

type subscription struct {
	handler func(topic string, payload []byte)
}

type Client struct {
	opts *mqtt.ClientOptions
	inner mqtt.Client
	mu   sync.Mutex
	subs map[string]subscription
}

func New(cfg *config.Config) (*Client, error) {
	c := &Client{subs: make(map[string]subscription)}

	c.opts = mqtt.NewClientOptions().
		AddBroker(cfg.MQTT.Broker).
		SetClientID(cfg.MQTT.ClientID).
		SetUsername(cfg.MQTT.Username).
		SetPassword(cfg.MQTT.Password).
		SetTLSConfig(&tls.Config{}).
		SetConnectTimeout(10 * time.Second).
		SetAutoReconnect(true).
		SetCleanSession(false).
		SetOnConnectHandler(func(inner mqtt.Client) {
			// Re-apply subscriptions on every (re)connect.
			// By the time this fires, c.subs is always populated because
			// Connect() is called only after Subscribe() has been called.
			c.mu.Lock()
			subs := make(map[string]subscription, len(c.subs))
			for k, v := range c.subs {
				subs[k] = v
			}
			c.mu.Unlock()

			for topic, sub := range subs {
				sub := sub
				inner.Subscribe(topic, 1, func(_ mqtt.Client, msg mqtt.Message) {
					sub.handler(msg.Topic(), msg.Payload())
				})
			}
		})

	return c, nil
}

// Subscribe registers a handler for the given topic before the connection is opened.
// Must be called before Connect() so queued messages are not dropped on first connect.
func (c *Client) Subscribe(topic string, handler func(topic string, payload []byte)) {
	c.mu.Lock()
	c.subs[topic] = subscription{handler: handler}
	c.mu.Unlock()
}

// Connect opens the broker connection. Call this after all Subscribe() calls.
func (c *Client) Connect() error {
	c.inner = mqtt.NewClient(c.opts)
	if tok := c.inner.Connect(); tok.Wait() && tok.Error() != nil {
		return fmt.Errorf("mqtt connect: %w", tok.Error())
	}
	return nil
}

// Disconnect cleanly closes the connection.
func (c *Client) Disconnect() {
	c.inner.Disconnect(500)
}
