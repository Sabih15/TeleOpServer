package middleware

import (
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

// RequestLogger is a Chi-compatible middleware that logs each request with zerolog.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		next.ServeHTTP(ww, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("request_id", chimiddleware.GetReqID(r.Context())).
			Int("status", ww.Status()).
			Dur("latency", time.Since(start)).
			Msg("request")
	})
}
