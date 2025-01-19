package http_server

import (
	"app/internal/config"
	httpMiddleware "app/internal/http-server/middleware"
	"app/internal/http-server/router"
	"app/internal/storage"
	"app/pkg/client/rabbitmq"
	"app/pkg/common/logging"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(
	ctx context.Context,
	pgClient *pgxpool.Pool,
	cfg *config.Config,
	amqpClient *rabbitmq.App,
) (*chi.Mux, error) {
	r := chi.NewRouter()

	storages, err := storage.New(ctx, pgClient)
	if err != nil {
		logging.L(ctx).Error("failed to initialize storage", err)
		return nil, err
	}

	httpMiddleware.RegisterMiddlewares(r, ctx, cfg, amqpClient)

	routes.RegisterRoutes(r, ctx, cfg, storages, pgClient, amqpClient)

	logging.L(ctx).Info("server prepared successfully")

	return r, nil
}
