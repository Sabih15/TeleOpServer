package gps

import (
	"github.com/google/wire"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewRepository, NewService, NewConsumer)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&GPSReading{}); err != nil {
		return err
	}
	return db.Exec(`
		SELECT create_hypertable('gps_readings', 'time', if_not_exists => TRUE)
	`).Error
}
