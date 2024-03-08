package app

import (
	"app/internal/config"
	clientHTTP "app/internal/http-server/handlers/client"
	loginHTTP "app/internal/http-server/handlers/login"
	refreshHTTP "app/internal/http-server/handlers/refresh-token"
	registerHTTP "app/internal/http-server/handlers/register"
	clientStorage "app/internal/storage/pgsql/client"
	accessTokenStorage "app/internal/storage/pgsql/oauth/access-token"
	refreshTokenStorage "app/internal/storage/pgsql/oauth/refresh-token"
	"app/internal/storage/pgsql/user"
	"app/pkg/client/pgsql"
	"app/pkg/common/logging"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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

	storage, err := user.New(ctx, pgClient)
	if err != nil {
		logging.L(ctx).Error("failed to init storage user", err)
		return err
	}

	storageClient, err := clientStorage.New(ctx, pgClient)
	if err != nil {
		logging.L(ctx).Error("failed to init storage access token", err)
		return err
	}

	storageAccessToken, err := accessTokenStorage.New(ctx, pgClient)
	if err != nil {
		logging.L(ctx).Error("failed to init storage access token", err)
		return err
	}

	storageRefreshToken, err := refreshTokenStorage.New(ctx, pgClient)
	if err != nil {
		logging.L(ctx).Error("failed to init storage refresh token", err)
		return err
	}

	logging.L(ctx).Info("router initializing")
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/oauth/registration",
		registerHTTP.New(ctx, storage),
	)

	r.Post("/oauth/login",
		loginHTTP.New(
			ctx,
			storage,
			storageAccessToken,
			storageRefreshToken,
			storageClient,
			a.cfg.Token,
		),
	)

	r.Post("/oauth/refresh-token",
		refreshHTTP.New(
			ctx,
			storageAccessToken,
			storageRefreshToken,
			storageClient,
			a.cfg.Token,
		),
	)

	var clnt = clientHTTP.New(ctx, storageClient)
	r.Get("/oauth/get-client/{client:[a-z]{1,20}}", clnt.GetClient())

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
