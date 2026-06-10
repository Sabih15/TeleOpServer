package TOCommands

import "time"

// TeleOpCommand is the hypertable model — no gorm.Model, Time is the partition key.
type TeleOpCommand struct {
	Time    time.Time `gorm:"not null;index"`
	RobotID uint      `gorm:"not null;index"`
	UserID  uint      `gorm:"not null"`
	Command string    `gorm:"not null"`
	MsgID   uint64    `gorm:"not null"`
	T1      uint64    `gorm:"not null"`
	Lx      float64   `gorm:"not null"`
	Ly      float64   `gorm:"not null"`
	Az      float64   `gorm:"not null"`
}

// CreateCommandRequest is the ingest payload — Time is set server-side.
type CreateCommandRequest struct {
	Time    time.Time `json:"time"`
	RobotID uint      `json:"robot_id"`
	UserID  uint      `json:"user_id"`
	Command string    `json:"command"`
	MsgID   uint64    `json:"msg_id"`
	T1      uint64    `json:"t1"`
	Lx      float64   `json:"lx"`
	Ly      float64   `json:"ly"`
	Az      float64   `json:"az"`
}

// CommandResponse is what the API returns.
type CommandResponse struct {
	Time    time.Time `json:"time"`
	RobotID uint      `json:"robot_id"`
	UserID  uint      `json:"user_id"`
	Command string    `json:"command"`
	MsgID   uint64    `json:"msg_id"`
	T1      uint64    `json:"t1"`
	Lx      float64   `json:"lx"`
	Ly      float64   `json:"ly"`
	Az      float64   `json:"az"`
}

func toCommandResponse(c *TeleOpCommand) CommandResponse {
	return CommandResponse{
		Time:    c.Time,
		RobotID: c.RobotID,
		UserID:  c.UserID,
		Command: c.Command,
		MsgID:   c.MsgID,
		T1:      c.T1,
		Lx:      c.Lx,
		Ly:      c.Ly,
		Az:      c.Az,
	}
}
