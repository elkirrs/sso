package refresh_token

import (
	refreshTokenDomain "app/internal/domain/oauth/refresh-token"
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

func (s *Storage) CreateRefreshToken(rT *refreshTokenDomain.RefreshToken) (string, error) {
	const op = "storage.pgsql.oauth.refresh-token.CreateRefreshToken"
	s.log.Info("op", op)

	querySQL := `
			INSERT INTO %s (id, access_token_id, revoked, expires_at)
			VALUES ($1, $2, $3, $4)
			RETURNING ID
			`
	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthRefreshToken)
	querySQL = loop.FormatQuery(querySQL)
	s.log.Info("query", querySQL)

	var refreshTokenStorage refreshTokenDomain.RefreshToken

	err := s.db.QueryRow(
		s.ctx,
		querySQL,
		rT.ID,
		rT.AccessTokenId,
		rT.Revoked,
		rT.ExpiresAt,
	).Scan(
		&refreshTokenStorage.ID,
	)

	if err != nil {
		s.log.Error("error query", err)
		return "", err
	}

	return refreshTokenStorage.ID, nil
}

func (s *Storage) ExistsToken(rT *refreshTokenDomain.RefreshToken) (bool, error) {
	const op = "storage.pgsql.oauth.refresh-token.ExistsToken"
	s.log.Info("op", op)
	var isExists bool
	querySQL := `
		SELECT (COUNT(*)::smallint)::int::bool as isExists FROM %s WHERE id = $1 AND access_token_id = $2
	`
	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthRefreshToken)
	querySQL = loop.FormatQuery(querySQL)
	s.log.Info("query", querySQL)

	err := s.db.QueryRow(
		s.ctx,
		querySQL,
		rT.ID,
		rT.AccessTokenId,
	).Scan(&isExists)

	if err != nil {
		s.log.Error("error query", err)
		return false, err
	}

	return isExists, nil
}

func (s *Storage) UpdateToken(rT *refreshTokenDomain.RefreshToken) (bool, error) {
	const op = "storage.pgsql.oauth.refresh-token.UpdateToken"
	s.log.Info("op", op)

	querySQL := `
		UPDATE %s SET revoked = true WHERE id = $1 AND access_token_id = $2
	`
	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthRefreshToken)
	querySQL = loop.FormatQuery(querySQL)
	s.log.Info("query", querySQL)
	_, err := s.db.Exec(
		s.ctx,
		querySQL,
		rT.ID,
		rT.AccessTokenId,
	)

	if err != nil {
		s.log.Error("error query", err)
		return false, err
	}

	return true, nil
}
