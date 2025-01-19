package routes

import (
	"app/internal/config"
	"app/internal/http-server/handlers/health"
	"app/pkg/client/rabbitmq"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterHealthRoutes(
	r chi.Router,
	ctx context.Context,
	cfg *config.Config,
	pgClient *pgxpool.Pool,
	amqpClient *rabbitmq.App,
) {
	r.Get("/health", health.New(ctx, cfg, pgClient, amqpClient))
}
