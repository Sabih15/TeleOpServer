package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

func (s *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf(":%s", s.cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: s.router}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
