package TOCommands

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/sabih15/TeleOpServer/internal/platform/mqttclient"
)

const commandTopic = "teleopserver/robots/+/commands"

type Consumer struct {
	mqtt    *mqttclient.Client
	service IService
}

func NewConsumer(mqtt *mqttclient.Client, service IService) *Consumer {
	return &Consumer{mqtt: mqtt, service: service}
}

// Start registers the subscription handler, then opens the broker connection so
// queued messages are never delivered before the handler is in place.
func (c *Consumer) Start(ctx context.Context) error {
	c.mqtt.Subscribe(commandTopic, func(topic string, payload []byte) {
		var req CreateCommandRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("mqtt: failed to parse command payload")
			return
		}

		if err := c.service.Record(ctx, req); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("mqtt: failed to record command")
			return
		}

		log.Debug().Str("topic", topic).Uint("robot_id", req.RobotID).Msg("mqtt: command recorded")
	})

	if err := c.mqtt.Connect(); err != nil {
		return err
	}

	<-ctx.Done()
	c.mqtt.Disconnect()
	return nil
}
