package gps

import (
	"context"

	"gorm.io/gorm"
)

type IRepository interface {
	Save(ctx context.Context, r *GPSReading) error
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
