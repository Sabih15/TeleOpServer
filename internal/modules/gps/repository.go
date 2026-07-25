package gps

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type IRepository interface {
	Save(ctx context.Context, r *GPSReading) error
	FindByRobotAndTimeRange(ctx context.Context, robotID uint, from, to time.Time) ([]GPSReading, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) IRepository {
	return &repository{db: db}
}

func (r *repository) Save(ctx context.Context, reading *GPSReading) error {
	return r.db.WithContext(ctx).Create(reading).Error
}

func (r *repository) FindByRobotAndTimeRange(ctx context.Context, robotID uint, from, to time.Time) ([]GPSReading, error) {
	var readings []GPSReading
	err := r.db.WithContext(ctx).
		Where("robot_id = ? AND time >= ? AND time <= ?", robotID, from, to).
		Order("time ASC").
		Find(&readings).Error
	return readings, err
}
