package gps

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/sabih15/TeleOpServer/internal/platform/mqttclient"
)

const gpsTopic = "teleopserver/robots/+/gps"

type Consumer struct {
	mqtt    *mqttclient.Client
	service IService
}

func NewConsumer(mqtt *mqttclient.Client, service IService) *Consumer {
	return &Consumer{mqtt: mqtt, service: service}
}

func (c *Consumer) Register() {
	c.mqtt.Subscribe(gpsTopic, func(topic string, payload []byte) {
		var req CreateGPSRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("mqtt: failed to parse gps payload")
			return
		}

		if err := c.service.Record(context.Background(), req); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("mqtt: failed to record gps reading")
			return
		}

		log.Debug().Str("topic", topic).Uint("robot_id", req.RobotID).Msg("mqtt: gps reading recorded")
	})
}
