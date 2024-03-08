package app

import (
	"app/internal/config"
	"app/internal/http-server/router"
	"app/pkg/client/pgsql"
	"app/pkg/common/logging"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/rs/cors"
	"net"
	"net/http"
)

type App struct {
	ctx        context.Context
	router     *chi.Mux
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
	const op = "app.Run"
	ctx := a.ctx
	logging.L(ctx).Info("op", op)

	logging.WithAttrs(ctx,
		logging.StringAttr("host", a.cfg.Host),
		logging.IntAttr("port", a.cfg.HTTP.Port),
	)

	pgClient, err := pgsql.New(ctx, a.cfg.DB)

	if err != nil {
		logging.L(ctx).Error("failed to connect to db", err)
		return err
	}

	logging.L(ctx).Info("DB connected")

	r, err := router.GetRouters(ctx, pgClient, a.cfg)

	if err != nil {
		logging.L(ctx).Error("failed to create routers", err)
		return err
	}

	logging.L(ctx).Info("starting api server")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.HTTP.Port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	c := cors.New(cors.Options{
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodOptions, http.MethodDelete},
		//AllowedOrigins:     []string{"http://localhost:3000", "http://localhost:80"},
		//AllowCredentials:   true,
		//AllowedHeaders:     []string{"Location", "Charset", "Access-Control-Allow-Origin", "Content-Type", "content-type", "Origin", "Accept", "Content-Length", "Accept-Encoding", "X-CSRF-Token"},
		//OptionsPassthrough: true,
		//ExposedHeaders:     []string{"Location", "Authorization", "Content-Disposition"},
		// Enable Debugging for testing, consider disabling in production
		//Debug: false,
	})

	handler := c.Handler(r)

	a.httpServer = &http.Server{
		Handler:      handler,
		WriteTimeout: a.cfg.HTTP.WriteTimeout,
		ReadTimeout:  a.cfg.HTTP.ReadTimeout,
	}

	if err := a.httpServer.Serve(l); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logging.L(ctx).Warn("server shutdown")
			return fmt.Errorf("%s: %w", op, err)
		default:
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := a.httpServer.Shutdown(ctx); err != nil {
		logging.L(ctx).Error("failed to stop server", err)
		return err
	}

	return nil
}
