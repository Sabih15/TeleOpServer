package user

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/wire"
	"github.com/sabih15/TeleOpServer/internal/platform/config"
	"github.com/sabih15/TeleOpServer/internal/platform/middleware"
	"gorm.io/gorm"
)

// ProviderSet groups all user module constructors for Wire.
var ProviderSet = wire.NewSet(NewRepository, NewService, NewHandler)

// Migrate creates/updates this module's own database tables.
// Called once at startup --- each module owns its own schema.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}

// RegisterRoutes mounts all user module endpoints onto the router.
// Called from platform/server so the server never needs to know handler internals.
func RegisterRoutes(r chi.Router, cfg *config.Config, h *Handler) {
	// Public
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)

	// Protected — JWT required
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg))
		r.Get("/users/me", h.GetProfile)
	})
}
