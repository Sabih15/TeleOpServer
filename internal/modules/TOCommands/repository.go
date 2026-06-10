package TOCommands

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type IRepository interface {
	Save(ctx context.Context, cmd *TeleOpCommand) error
	FindByRobotAndTimeRange(ctx context.Context, robotID uint, from, to time.Time) ([]TeleOpCommand, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) IRepository {
	return &repository{db: db}
}

func (r *repository) Save(ctx context.Context, cmd *TeleOpCommand) error {
	return r.db.WithContext(ctx).Create(cmd).Error
}

func (r *repository) FindByRobotAndTimeRange(ctx context.Context, robotID uint, from, to time.Time) ([]TeleOpCommand, error) {
	var cmds []TeleOpCommand
	err := r.db.WithContext(ctx).
		Where("robot_id = ? AND time >= ? AND time <= ?", robotID, from, to).
		Order("time ASC").
		Find(&cmds).Error
	return cmds, err
}
