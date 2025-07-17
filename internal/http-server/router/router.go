package routes

import (
	"app/internal/config"
	"app/internal/storage"
	"app/pkg/client/rabbitmq"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(
	r chi.Router,
	ctx context.Context,
	cfg *config.Config,
	storages *storage.Storage,
	pgClient *pgxpool.Pool,
	queueClient *rabbitmq.App,
) {
	RegisterOAuthRoutes(r, ctx, storages, cfg)
	RegisterHealthRoutes(r, ctx, cfg, pgClient, queueClient)
}
