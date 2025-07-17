package middleware

import (
	"app/internal/config"
	"app/pkg/client/rabbitmq"
	"app/pkg/common/logging"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func RegisterMiddlewares(
	r chi.Router,
	ctx context.Context,
	cfg *config.Config,
	queueClient *rabbitmq.App,
) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(MetricsPrometheus)

	if cfg.Queue.Driver != "" {
		r.Use(Logging(ctx, queueClient))
	}

	logging.L(ctx).Info("Middleware initialized successfully")
}
