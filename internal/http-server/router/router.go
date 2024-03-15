package router

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
	"app/pkg/common/logging"
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetRouters(ctx context.Context, pgClient *pgxpool.Pool, cfg *config.Config) (*chi.Mux, error) {
	r := chi.NewRouter()

	storageUser, err := user.New(ctx, pgClient)
	if err != nil {
		logging.L(ctx).Error("failed to init storage user", err)
		return r, err
	}

	storageClient, err := clientStorage.New(ctx, pgClient)
	if err != nil {
		logging.L(ctx).Error("failed to init storage access token", err)
		return r, err
	}

	storageAccessToken, err := accessTokenStorage.New(ctx, pgClient)
	if err != nil {
		logging.L(ctx).Error("failed to init storage access token", err)
		return r, err
	}

	storageRefreshToken, err := refreshTokenStorage.New(ctx, pgClient)
	if err != nil {
		logging.L(ctx).Error("failed to init storage refresh token", err)
		return r, err
	}

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/oauth/registration",
		registerHTTP.New(ctx, storageUser),
	)

	r.Post("/oauth/login",
		loginHTTP.New(
			ctx,
			storageUser,
			storageAccessToken,
			storageRefreshToken,
			storageClient,
			cfg.Token,
		),
	)

	r.Post("/oauth/refresh-token",
		refreshHTTP.New(
			ctx,
			storageAccessToken,
			storageRefreshToken,
			storageClient,
			cfg.Token,
		),
	)

	var client = clientHTTP.New(ctx, storageClient)
	r.Get("/oauth/client/{client:[a-z]{1,20}}", client.GetClient())
	r.Post("/oauth/client", client.CreateClient())

	return r, nil
}
