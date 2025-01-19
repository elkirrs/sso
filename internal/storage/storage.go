package storage

import (
	clientStorage "app/internal/storage/pgsql/client"
	accessToken "app/internal/storage/pgsql/oauth/access-token"
	refreshToken "app/internal/storage/pgsql/oauth/refresh-token"
	"app/internal/storage/pgsql/user"
	"app/pkg/common/logging"
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCodeExists = "23505"
)

func ErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}

	return ""
}

type Storage struct {
	User         *user.Storage
	Client       *clientStorage.Storage
	AccessToken  *accessToken.Storage
	RefreshToken *refreshToken.Storage
}

func New(ctx context.Context, pgClient *pgxpool.Pool) (*Storage, error) {
	storageUser, err := user.New(ctx, pgClient)
	if err != nil {
		logging.L(ctx).Error("failed to init storage user", err)
		return nil, err
	}

	storageClient, err := clientStorage.New(ctx, pgClient)
	if err != nil {
		logging.L(ctx).Error("failed to init storage client", err)
		return nil, err
	}

	storageAccessToken, err := accessToken.New(ctx, pgClient)
	if err != nil {
		logging.L(ctx).Error("failed to init storage access token", err)
		return nil, err
	}

	storageRefreshToken, err := refreshToken.New(ctx, pgClient)
	if err != nil {
		logging.L(ctx).Error("failed to init storage refresh token", err)
		return nil, err
	}

	return &Storage{
		User:         storageUser,
		Client:       storageClient,
		AccessToken:  storageAccessToken,
		RefreshToken: storageRefreshToken,
	}, nil
}
