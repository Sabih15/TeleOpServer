package playback

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/wire"
	"github.com/sabih15/TeleOpServer/internal/platform/config"
	"github.com/sabih15/TeleOpServer/internal/platform/middleware"
	"gorm.io/gorm"
)

// ProviderSet groups all playback module constructors for Wire.
var ProviderSet = wire.NewSet(NewStorage, NewRepository, NewService, NewHandler, NewConsumer)

// Migrate creates the playback_frames table and converts it to a TimescaleDB hypertable.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&PlaybackFrame{}); err != nil {
		return err
	}
	return db.Exec(`
		SELECT create_hypertable('playback_frames', 'time', if_not_exists => TRUE)
	`).Error
}

// RegisterRoutes mounts all playback endpoints onto the router.
func RegisterRoutes(r chi.Router, cfg *config.Config, h *Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg))
		r.Get("/playback", h.GetHistory)
	})
}
