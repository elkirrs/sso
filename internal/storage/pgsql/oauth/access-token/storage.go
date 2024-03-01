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

func (s *Storage) ExistsToken(aT *accessToken.AccessToken) (bool, error) {
	const op = "storage.pgsql.oauth.access-token.ExistsToken"
	s.log.Info("op", op)
	var isExists bool
	querySQL := `
		SELECT (COUNT(*)::smallint)::int::bool as isExists FROM %s WHERE id = $1 AND user_id = $2 AND client_id = $3
	`
	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthAccessToken)
	querySQL = loop.FormatQuery(querySQL)
	s.log.Info("query", querySQL)

	err := s.db.QueryRow(
		s.ctx,
		querySQL,
		aT.ID,
		aT.UserId,
		aT.ClientId,
	).Scan(&isExists)

	if err != nil {
		s.log.Error("error query", err)
		return false, err
	}

	return isExists, nil
}

func (s *Storage) UpdateToken(aT *accessToken.AccessToken) (bool, error) {
	const op = "storage.pgsql.oauth.access-token.UpdateToken"
	s.log.Info("op", op)

	querySQL := `
		UPDATE %s SET revoked = true WHERE id = $1 AND user_id = $2 AND client_id = $3
	`
	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthAccessToken)
	querySQL = loop.FormatQuery(querySQL)
	s.log.Info("query", querySQL)
	_, err := s.db.Exec(
		s.ctx,
		querySQL,
		aT.ID,
		aT.UserId,
		aT.ClientId,
	)

	if err != nil {
		s.log.Error("error query", err)
		return false, err
	}

	return true, nil
}
