package TOCommands

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/wire"
	"github.com/sabih15/TeleOpServer/internal/platform/config"
	"github.com/sabih15/TeleOpServer/internal/platform/middleware"
	"gorm.io/gorm"
)

// ProviderSet groups all TOCommands module constructors for Wire.
var ProviderSet = wire.NewSet(NewRepository, NewService, NewHandler)

// Migrate creates the teleop_commands table and converts it to a TimescaleDB hypertable.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&TeleOpCommand{}); err != nil {
		return err
	}
	return db.Exec(`
		SELECT create_hypertable('tele_op_commands', 'time', if_not_exists => TRUE)
	`).Error
}

// RegisterRoutes mounts all TOCommands endpoints onto the router.
func RegisterRoutes(r chi.Router, cfg *config.Config, h *Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg))
		r.Post("/commands", h.RecordCommand)
		r.Get("/commands", h.GetHistory)
	})
}
