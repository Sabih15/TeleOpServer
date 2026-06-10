// wire/providers.go  — no build tag, always compiled
// This is the composition root — the one place that knows about all modules.
// Add new module migrations to MigrateAll and routes to provideRouter.
package wire

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	tocommands "github.com/sabih15/TeleOpServer/internal/modules/TOCommands"
	usermod "github.com/sabih15/TeleOpServer/internal/modules/user"
	"github.com/sabih15/TeleOpServer/internal/platform/config"
	"github.com/sabih15/TeleOpServer/internal/platform/mqttclient"
	"github.com/sabih15/TeleOpServer/internal/platform/server"
	"gorm.io/gorm"
)

// MigrateAll runs every module's migration in order.
// Add a new module's Migrate() call here when you create one.
func MigrateAll(db *gorm.DB) error {
	if err := usermod.Migrate(db); err != nil {
		return fmt.Errorf("user migration: %w", err)
	}
	if err := tocommands.Migrate(db); err != nil {
		return fmt.Errorf("tocommands migration: %w", err)
	}
	return nil
}

// provideRouter runs migrations then builds the fully configured Chi router.
// Wire injects *gorm.DB automatically from database.NewPostgres.
func provideRouter(cfg *config.Config, db *gorm.DB, mqtt *mqttclient.Client, userHandler *usermod.Handler, cmdHandler *tocommands.Handler) (*chi.Mux, error) {
	if err := MigrateAll(db); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	r := server.NewBaseRouter()
	r.Route("/api/v1", func(r chi.Router) {
		usermod.RegisterRoutes(r, cfg, userHandler)
		tocommands.RegisterRoutes(r, cfg, cmdHandler)
	})
	return r, nil
}
