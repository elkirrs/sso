package routes

import (
	"app/internal/config"
	clientHTTP "app/internal/http-server/handlers/client"
	loginHTTP "app/internal/http-server/handlers/login"
	refreshHTTP "app/internal/http-server/handlers/refresh-token"
	registerHTTP "app/internal/http-server/handlers/register"
	"app/internal/storage"
	"context"
	"github.com/go-chi/chi/v5"
)

func RegisterOAuthRoutes(
	r chi.Router,
	ctx context.Context,
	storages *storage.Storage,
	cfg *config.Config,
) {
	r.Post("/oauth/registration",
		registerHTTP.New(ctx, storages.User),
	)

	r.Post("/oauth/login",
		loginHTTP.New(
			ctx,
			storages.User,
			storages.AuthToken,
			storages.Client,
			cfg.Token,
		),
	)

	r.Post("/oauth/refresh-token",
		refreshHTTP.New(
			ctx,
			storages.AccessToken,
			storages.RefreshToken,
			storages.Client,
			cfg.Token,
		),
	)

	client := clientHTTP.New(ctx, storages.Client)
	r.Get("/oauth/client/{client:[a-z]{1,20}}", client.GetClient())
	r.Post("/oauth/client", client.CreateClient())
}
