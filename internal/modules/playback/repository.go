package playback

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type IRepository interface {
	Save(ctx context.Context, f *PlaybackFrame) error
	FindByRobotAndTimeRange(ctx context.Context, robotID uint, from, to time.Time) ([]PlaybackFrame, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) IRepository {
	return &repository{db: db}
}

func (r *repository) Save(ctx context.Context, frame *PlaybackFrame) error {
	return r.db.WithContext(ctx).Create(frame).Error
}

func (r *repository) FindByRobotAndTimeRange(ctx context.Context, robotID uint, from, to time.Time) ([]PlaybackFrame, error) {
	var frames []PlaybackFrame
	err := r.db.WithContext(ctx).
		Where("robot_id = ? AND time >= ? AND time <= ?", robotID, from, to).
		Order("time ASC").
		Find(&frames).Error
	return frames, err
}
