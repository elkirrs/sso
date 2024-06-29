package app

import (
	"app/internal/config"
	"app/pkg/common/logging"
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type App struct {
	ctx        context.Context
	httpServer *http.Server
	cfg        *config.Config
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
	const op = "app.metrics.Run"
	logging.L(a.ctx).Info("op", op)

	logging.WithAttrs(a.ctx,
		logging.StringAttr("host", a.cfg.Host),
		logging.IntAttr("port", a.cfg.Metrics.Port),
	)

	logging.L(a.ctx).Info("starting metrics server")

	host := fmt.Sprintf("%s:%d", a.cfg.Host, a.cfg.Metrics.Port)

	a.httpServer = &http.Server{}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	logging.L(a.ctx).Info("http metrics is running", logging.StringAttr("addr", host))

	err := http.ListenAndServe(host, mux)

	if err != nil {
		logging.L(a.ctx).Error("metrics server", err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "app.metrics.Stop"
	logging.L(a.ctx).Info("op", op)

	if err := a.httpServer.Shutdown(a.ctx); err != nil {
		logging.L(a.ctx).Error("failed to stop metrics server", err)
		return
	}

	logging.L(a.ctx).Info("metrics server stopped")
}
