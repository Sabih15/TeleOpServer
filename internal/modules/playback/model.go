package playback

import "time"

// PlaybackFrame is the hypertable model — no gorm.Model, Time is the partition key.
// The frame bytes themselves live on disk; this row is metadata + a pointer to them.
type PlaybackFrame struct {
	Time        time.Time `gorm:"not null;index"`
	RobotID     uint      `gorm:"not null;index"`
	StoragePath string    `gorm:"not null"`
	SizeBytes   int64     `gorm:"not null"`
}

// CreatePlaybackRequest is the MQTT ingest payload. RobotID is not present here —
// it's parsed from the topic instead. Data is base64 in the JSON wire format;
// encoding/json decodes it to raw bytes automatically for a []byte field.
type CreatePlaybackRequest struct {
	Time time.Time `json:"time"`
	Data []byte    `json:"data"`
}

// PlaybackFrameResponse is what the API returns.
type PlaybackFrameResponse struct {
	Time        time.Time `json:"time"`
	RobotID     uint      `json:"robot_id"`
	StoragePath string    `json:"storage_path"`
	SizeBytes   int64     `json:"size_bytes"`
}

func toPlaybackFrameResponse(f *PlaybackFrame) PlaybackFrameResponse {
	return PlaybackFrameResponse{
		Time:        f.Time,
		RobotID:     f.RobotID,
		StoragePath: f.StoragePath,
		SizeBytes:   f.SizeBytes,
	}
}
