package app

import (
	appGRPC "app/internal/app/grpc"
	appApi "app/internal/app/http"
	appMetrics "app/internal/app/metrics"
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
	pgClient         *pgxpool.Pool
	metricsServerApp *appMetrics.App
	amqpClient       *rabbitmq.App
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

	pgClient, err := pgsql.New(a.ctx, a.cfg.DB)

	if err != nil {
		logging.L(a.ctx).Error("failed to connect to db", err)
		return err
	}

	a.pgClient = pgClient
	logging.L(a.ctx).Info("DB connected")

	if a.cfg.Queue.Driver != "" {
		a.amqpClient = rabbitmq.New(a.ctx, a.cfg)
		err = a.amqpClient.Connect()
		if err != nil {
			logging.L(a.ctx).Error("Couldn't connect to RabbitMQ: ", err)
			return err
		}
		err = a.amqpClient.SetupQueueAndExchange("logs")

		if err != nil {
			logging.L(a.ctx).Error("Error during setup RabbitMQ: ", err)
			return err
		}
	}

	a.httpServerApp = appApi.New(a.ctx, pgClient, a.cfg, a.amqpClient)
	a.gRPCServerApp = appGRPC.New(a.ctx, pgClient, a.cfg)
	a.metricsServerApp = appMetrics.New(a.ctx, a.cfg)

	go a.httpServerApp.MustRun()
	go a.gRPCServerApp.MustRun()
	go a.metricsServerApp.MustRun()

	return nil
}

func (a *App) Stop() {
	const op = "app.Stop"
	logging.L(a.ctx).Info("op", op)

	a.httpServerApp.Stop()
	a.gRPCServerApp.Stop()
	a.metricsServerApp.Stop()

	if a.pgClient != nil {
		a.pgClient.Close()
		logging.L(a.ctx).Info("Connection to DB closed")
	}

	a.amqpClient.Close()
	logging.L(a.ctx).Info("Connection to RabbitMQ closed")
}
