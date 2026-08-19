package playback

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/sabih15/TeleOpServer/internal/platform/mqttclient"
)

const playbackTopic = "teleopserver/robots/+/playback"

type Consumer struct {
	mqtt    *mqttclient.Client
	service IService
}

func NewConsumer(mqtt *mqttclient.Client, service IService) *Consumer {
	return &Consumer{mqtt: mqtt, service: service}
}

// Register subscribes to the playback topic. The payload is JSON carrying the
// source timestamp and the frame as base64; robot_id is parsed from the
// concrete topic instead, same as the other consumers key off it.
func (c *Consumer) Register() {
	c.mqtt.Subscribe(playbackTopic, func(topic string, payload []byte) {
		robotID, err := robotIDFromTopic(topic)
		if err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("mqtt: failed to parse playback topic")
			return
		}

		var req CreatePlaybackRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("mqtt: failed to parse playback payload")
			return
		}

		if err := c.service.Record(context.Background(), robotID, req.Time, req.Data); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("mqtt: failed to record playback frame")
			return
		}

		log.Debug().Str("topic", topic).Uint("robot_id", robotID).Int("bytes", len(req.Data)).Msg("mqtt: playback frame recorded")
	})
}

// robotIDFromTopic extracts the robot ID from a concrete topic matching
// teleopserver/robots/<id>/playback.
func robotIDFromTopic(topic string) (uint, error) {
	parts := strings.Split(topic, "/")
	if len(parts) != 4 {
		return 0, fmt.Errorf("unexpected topic format: %s", topic)
	}

	id, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid robot id in topic %q: %w", topic, err)
	}
	return uint(id), nil
}
