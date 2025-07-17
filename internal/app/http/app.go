package app

import (
	"app/internal/config"
	server "app/internal/http-server"
	"app/pkg/client/rabbitmq"
	"app/pkg/common/logging"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"
	"net"
	"net/http"
)

type App struct {
	ctx         context.Context
	router      *chi.Mux
	httpServer  *http.Server
	cfg         *config.Config
	pgClient    *pgxpool.Pool
	queueClient *rabbitmq.App
}

func New(
	ctx context.Context,
	pgClient *pgxpool.Pool,
	cfg *config.Config,
	queueClient *rabbitmq.App,
) *App {
	return &App{
		ctx:         ctx,
		cfg:         cfg,
		pgClient:    pgClient,
		queueClient: queueClient,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "app.http.Run"
	logging.L(a.ctx).Info("op", op)

	logging.WithAttrs(a.ctx,
		logging.StringAttr("host", a.cfg.Host),
		logging.IntAttr("port", a.cfg.HTTP.Port),
	)

	r, err := server.New(a.ctx, a.pgClient, a.cfg, a.queueClient)

	if err != nil {
		logging.L(a.ctx).Error("failed to create routers", err)
		return err
	}

	logging.L(a.ctx).Info("starting http server")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.HTTP.Port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	c := cors.New(cors.Options{
		AllowedMethods:     a.cfg.HTTP.CORS.AllowedMethods,
		AllowedOrigins:     a.cfg.HTTP.CORS.AllowedOrigins,
		AllowCredentials:   a.cfg.HTTP.CORS.AllowCredentials,
		AllowedHeaders:     a.cfg.HTTP.CORS.AllowedHeaders,
		OptionsPassthrough: a.cfg.HTTP.CORS.OptionsPassthrough,
		//ExposedHeaders:     a.cfg.HTTP.CORS.ExposedHeaders,
		// Enable Debugging for testing, consider disabling in production
		Debug: a.cfg.HTTP.CORS.Debug,
	})

	handler := c.Handler(r)

	a.httpServer = &http.Server{
		Handler:      handler,
		WriteTimeout: a.cfg.HTTP.WriteTimeout,
		ReadTimeout:  a.cfg.HTTP.ReadTimeout,
	}

	logging.L(a.ctx).Info("http server is running", logging.StringAttr("addr", l.Addr().String()))

	if err := a.httpServer.Serve(l); err != nil {
		logging.L(a.ctx).Error("http server", err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "app.http.Stop"
	logging.L(a.ctx).Info("op", op)

	if err := a.httpServer.Shutdown(a.ctx); err != nil {
		logging.L(a.ctx).Error("failed to stop http server", err)
		return
	}

	logging.L(a.ctx).Info("http server stopped")
}
