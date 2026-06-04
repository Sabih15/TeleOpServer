package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/sabih15/TeleOpServer/internal/platform/config"
	appmiddleware "github.com/sabih15/TeleOpServer/internal/platform/middleware"
	httpswagger "github.com/swaggo/http-swagger"

	_ "github.com/sabih15/TeleOpServer/docs" // registers the generated swagger spec
)

type Server struct {
	router *chi.Mux
	cfg    *config.Config
}

func NewBaseRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(appmiddleware.RequestLogger)

	r.Get("/swagger/*", httpswagger.WrapHandler)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	return r
}

func NewServer(cfg *config.Config, r *chi.Mux) *Server {
	return &Server{router: r, cfg: cfg}
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.cfg.Server.Port)
	return http.ListenAndServe(addr, s.router)
}
