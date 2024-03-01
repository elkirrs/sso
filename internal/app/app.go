package app

import (
	"app/internal/config"
	"app/internal/http-server/handlers/login"
	refresh "app/internal/http-server/handlers/refresh-token"
	"app/internal/http-server/handlers/register"
	accessTokenStorage "app/internal/storage/pgsql/oauth/access-token"
	"app/internal/storage/pgsql/oauth/client"
	refreshTokenStorage "app/internal/storage/pgsql/oauth/refresh-token"
	"app/internal/storage/pgsql/user"
	"app/pkg/client/pgsql"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"log/slog"
	"net"
	"net/http"
)

type App struct {
	log        *slog.Logger
	router     *chi.Mux
	httpServer *http.Server
	cfg        *config.Config
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	return &App{
		cfg: cfg,
		log: log,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "app.Run"

	ctx := context.Background()

	a.log.With(
		slog.String("op", op),
		slog.String("host", a.cfg.Host),
		slog.Int("port", a.cfg.HTTP.Port),
	)

	pgClient, err := pgsql.New(ctx, a.cfg.DB, a.log)

	if err != nil {
		a.log.Error("failed to connect to db", err)
		return err
	}

	a.log.Info("DB connected")

	storage, err := user.New(pgClient, a.log, ctx)
	if err != nil {
		a.log.Error("failed to init storage user", err)
		return err
	}

	storageClient, err := client.New(ctx, pgClient, a.log)
	if err != nil {
		a.log.Error("failed to init storage access token", err)
		return err
	}

	storageAccessToken, err := accessTokenStorage.New(ctx, pgClient, a.log)
	if err != nil {
		a.log.Error("failed to init storage access token", err)
		return err
	}

	storageRefreshToken, err := refreshTokenStorage.New(ctx, pgClient, a.log)
	if err != nil {
		a.log.Error("failed to init storage refresh token", err)
		return err
	}

	a.log.Info("router initializing")
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/oauth/registration",
		register.New(a.log, storage),
	)

	r.Post("/oauth/login",
		login.New(
			a.log,
			storage,
			storageAccessToken,
			storageRefreshToken,
			storageClient,
			a.cfg.Token,
		),
	)

	r.Post("/oauth/refresh-token",
		refresh.New(
			a.log,
			storageAccessToken,
			storageRefreshToken,
			storageClient,
			a.cfg.Token,
		),
	)

	a.log.Info("starting api server")

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
			a.log.Warn("server shutdown")
			return fmt.Errorf("%s: %w", op, err)
		default:
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.log.Error("failed to stop server", err)
		return err
	}

	return nil
}
