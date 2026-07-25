package gps

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/wire"
	"github.com/sabih15/TeleOpServer/internal/platform/config"
	"github.com/sabih15/TeleOpServer/internal/platform/middleware"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewRepository, NewService, NewHandler, NewConsumer)

func RegisterRoutes(r chi.Router, cfg *config.Config, h *Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg))
		r.Get("/gps", h.GetHistory)
	})
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&GPSReading{}); err != nil {
		return err
	}
	return db.Exec(`
		SELECT create_hypertable('gps_readings', 'time', if_not_exists => TRUE)
	`).Error
}
