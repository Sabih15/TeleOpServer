package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/sabih15/TeleOpServer/internal/config"
	"github.com/sabih15/TeleOpServer/internal/domain/user"
	appmiddleware "github.com/sabih15/TeleOpServer/internal/middleware"
	httpswagger "github.com/swaggo/http-swagger"

	_ "github.com/sabih15/TeleOpServer/docs" // registers the generated swagger spec
)

type Server struct {
	router *chi.Mux
	cfg    *config.Config
}

// NewServer wires the Chi router, registers all middleware and routes, and returns a ready Server.
// Wire injects *config.Config and *user.Handler automatically.
// To refresh Swagger docs after editing handler annotations: swag init -g cmd/api/main.go -o docs
func NewServer(cfg *config.Config, userHandler *user.Handler) *Server {
	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(appmiddleware.RequestLogger)

	// Swagger UI — visit /swagger/index.html
	r.Get("/swagger/*", httpswagger.WrapHandler)

	// Redirect root to Swagger UI
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Post("/auth/register", userHandler.Register)
		r.Post("/auth/login", userHandler.Login)

		// Protected routes — JWT required
		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.Auth(cfg))
			r.Get("/users/me", userHandler.GetProfile)
		})
	})

	return &Server{router: r, cfg: cfg}
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.cfg.Server.Port)
	return http.ListenAndServe(addr, s.router)
}
