package access_token

import (
	accessToken "app/internal/domain/oauth/access-token"
	"app/migrations"
	"app/pkg/utils/loop"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type Storage struct {
	db  *pgxpool.Pool
	log *slog.Logger
	ctx context.Context
}

func New(ctx context.Context, pgClient *pgxpool.Pool, log *slog.Logger) (*Storage, error) {
	return &Storage{
		db:  pgClient,
		log: log,
		ctx: ctx,
	}, nil
}

func (s *Storage) CreateToken(aT *accessToken.AccessToken) (string, error) {
	const op = "storage.pgsql.oauth.access-token.CreateToken"
	s.log.Info("op", op)

	querySQL := `
			INSERT INTO %s (id, user_id, client_id, name, scopes, revoked, created_at, updated_at, expires_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING ID
			`
	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthAccessToken)
	querySQL = loop.FormatQuery(querySQL)
	s.log.Info("query", querySQL)

	var AccessTokenStorage accessToken.AccessToken

	err := s.db.QueryRow(
		s.ctx,
		querySQL,
		aT.ID,
		aT.UserId,
		aT.ClientId,
		aT.Name,
		aT.Scopes,
		aT.Revoked,
		aT.CreatedAt,
		aT.UpdatedAt,
		aT.ExpiresAt,
	).Scan(
		&AccessTokenStorage.ID,
	)

	if err != nil {
		s.log.Error("error query", err)
		return "", err
	}

	return AccessTokenStorage.ID, nil
}
