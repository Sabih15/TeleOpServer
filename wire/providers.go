// wire/providers.go  — no build tag, always compiled
package wire

import (
	"github.com/go-chi/chi/v5"
	usermod "github.com/sabih15/TeleOpServer/internal/modules/user"
	"github.com/sabih15/TeleOpServer/internal/platform/config"
	"github.com/sabih15/TeleOpServer/internal/platform/server"
)

func provideRouter(cfg *config.Config, userHandler *usermod.Handler) *chi.Mux {
	r := server.NewBaseRouter()
	r.Route("/api/v1", func(r chi.Router) {
		usermod.RegisterRoutes(r, cfg, userHandler)
	})
	return r
}
