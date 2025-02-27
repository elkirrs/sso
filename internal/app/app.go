package app

import (
	appGRPC "app/internal/app/grpc"
	appApi "app/internal/app/http"
	appMetrics "app/internal/app/metrics"
	appQueue "app/internal/app/queue"
	"app/internal/config"
	"app/pkg/client/pgsql"
	"app/pkg/client/rabbitmq"
	"app/pkg/common/logging"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	ctx              context.Context
	router           *chi.Mux
	httpServerApp    *appApi.App
	gRPCServerApp    *appGRPC.App
	cfg              *config.Config
	dbClient         *pgxpool.Pool
	metricsServerApp *appMetrics.App
	queueClient      *rabbitmq.App
	queueApp         *appQueue.App
}

func New(
	ctx context.Context,
	cfg *config.Config,
) *App {
	return &App{
		ctx: ctx,
		cfg: cfg,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "app.Run"
	logging.L(a.ctx).Info("op", op)

	logging.WithAttrs(a.ctx,
		logging.StringAttr("host", a.cfg.Host),
		logging.IntAttr("port", a.cfg.HTTP.Port),
	)

	dbClient, err := pgsql.New(a.ctx, a.cfg.DB)

	if err != nil {
		logging.L(a.ctx).Error("failed to connect to db", err)
		return err
	}

	a.dbClient = dbClient
	logging.L(a.ctx).Info("DB connected")

	queueClient, err := rabbitmq.New(a.ctx, a.cfg.Queue)

	if err != nil {
		logging.L(a.ctx).Error("failed to connect to queue", err)
		return err
	}

	a.queueClient = queueClient
	logging.L(a.ctx).Info("Queue connected")

	a.httpServerApp = appApi.New(a.ctx, dbClient, a.cfg, queueClient)
	a.gRPCServerApp = appGRPC.New(a.ctx, dbClient, a.cfg)
	a.metricsServerApp = appMetrics.New(a.ctx, a.cfg)
	a.queueApp = appQueue.New(a.ctx, a.cfg, queueClient, dbClient)

	go a.httpServerApp.MustRun()
	go a.gRPCServerApp.MustRun()
	go a.metricsServerApp.MustRun()
	go a.queueApp.MustRun()

	return nil
}

func (a *App) Stop() {
	const op = "app.Stop"
	logging.L(a.ctx).Info("op", op)

	a.httpServerApp.Stop()
	a.gRPCServerApp.Stop()
	a.metricsServerApp.Stop()

	if a.dbClient != nil {
		a.dbClient.Close()
		logging.L(a.ctx).Info("Connection to DB closed")
	}

	a.queueClient.Close()
	logging.L(a.ctx).Info("Connection to RabbitMQ closed")
}
