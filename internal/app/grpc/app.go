package app

import (
	"app/internal/config"
	"app/internal/grpc-server/handler/client"
	"app/pkg/common/logging"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"net"
)

type App struct {
	ctx        context.Context
	cfg        *config.Config
	pgClient   *pgxpool.Pool
	gRPCServer *grpc.Server
}

func New(
	ctx context.Context,
	pgClient *pgxpool.Pool,
	cfg *config.Config,
) *App {
	gRPCServer := grpc.NewServer()
	return &App{
		ctx:        ctx,
		cfg:        cfg,
		pgClient:   pgClient,
		gRPCServer: gRPCServer,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "app.grpc.Run"

	logging.L(a.ctx).Info("op", op)

	logging.WithAttrs(a.ctx,
		logging.StringAttr("host", a.cfg.Host),
		logging.IntAttr("port", a.cfg.GRPC.Port),
	)

	logging.L(a.ctx).Info("starting gRPC server")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.GRPC.Port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	logging.L(a.ctx).Info("gRPC server is running", logging.StringAttr("addr", l.Addr().String()))

	client.Register(a.ctx, a.gRPCServer, a.pgClient)
	//registration.Register(a.ctx, a.gRPCServer, a.pgClient)

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "app.grpc.Stop"
	logging.L(a.ctx).Info("op", op)

	a.gRPCServer.GracefulStop()
	logging.L(a.ctx).Info("grpc server stopped")
}
