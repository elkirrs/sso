package app

import (
	"app/internal/config"
	"app/internal/queue/handlers"
	"app/internal/storage"
	"app/pkg/client/rabbitmq"
	"app/pkg/common/logging"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	ctx         context.Context
	cfg         *config.Config
	queueClient *rabbitmq.App
	dbClient    *pgxpool.Pool
}

func New(
	ctx context.Context,
	cfg *config.Config,
	queueClient *rabbitmq.App,
	dbClient *pgxpool.Pool,
) *App {
	return &App{
		ctx:         ctx,
		cfg:         cfg,
		queueClient: queueClient,
		dbClient:    dbClient,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "app.queue.Run"
	logging.L(a.ctx).Info("op", op)

	storages, err := storage.New(a.ctx, a.dbClient)
	if err != nil {
		logging.L(a.ctx).Error("failed to initialize storage", err)
		return err
	}

	registrationHandler := handlers.NewHandleRegistration(a.dbClient, storages)

	go a.queueClient.ConsumeMsg("sso:user-registration", func(msg amqp.Delivery) {
		rabbitmq.ProcessMessage(a.ctx, msg, registrationHandler)
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	a.Stop()
	return nil
}

func (a *App) Stop() {
	const op = "app.queue.Stop"
	logging.L(a.ctx).Info("op", op)

	//err := a.queueClient.Close()
	//if err != nil {
	//	logging.L(a.ctx).Error("failed to stop queue client", err)
	//	return
	//}

	logging.L(a.ctx).Info("stopped listening queue")
}
