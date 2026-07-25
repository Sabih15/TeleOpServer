package gps

import "time"

type GPSReading struct {
	Time      time.Time `gorm:"not null;index"`
	RobotID   uint      `gorm:"not null;index"`
	Latitude  float64   `gorm:"not null"`
	Longitude float64   `gorm:"not null"`
}

type CreateGPSRequest struct {
	RobotID   uint    `json:"robot_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type GPSResponse struct {
	Time      time.Time `json:"time"`
	RobotID   uint      `json:"robot_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
}

func toGPSResponse(r *GPSReading) GPSResponse {
	return GPSResponse{
		Time:      r.Time,
		RobotID:   r.RobotID,
		Latitude:  r.Latitude,
		Longitude: r.Longitude,
	}
}
