package middleware

import (
	"app/internal/metrics"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

func MetricsPrometheus(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			metrics.ObserveHttpRequest(time.Since(time.Now()), ww.Status())
		}()

		next.ServeHTTP(ww, r)
	}

	return http.HandlerFunc(fn)
}
