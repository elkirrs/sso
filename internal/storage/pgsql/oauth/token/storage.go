package token

import (
	accessToken "app/internal/domain/oauth/access-token"
	refreshTokenDomain "app/internal/domain/oauth/refresh-token"
	"app/migrations"
	"app/pkg/common/logging"
	"app/pkg/utils/loop"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func New(ctx context.Context, pgClient *pgxpool.Pool) (*Storage, error) {
	return &Storage{
		ctx: ctx,
		db:  pgClient,
	}, nil
}

func (s *Storage) Create(
	aT *accessToken.AccessToken,
	rT *refreshTokenDomain.RefreshToken,
) error {
	const op = "storage.pgsql.oauth.token.Create"
	logging.L(s.ctx).Info("op", op)
	querySQL := `
		WITH inserted_access AS (
			INSERT INTO %s (id, user_id, client_id, NAME, scopes, revoked, created_at, updated_at, expires_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				RETURNING id),
			 inserted_refresh AS (
				 INSERT INTO %s (id, access_token_id, revoked, expires_at)
					 SELECT $10, id, $11, $12 FROM inserted_access)
		SELECT id
		FROM inserted_access
	`

	query := fmt.Sprintf(querySQL, migrations.TableOauthAccessToken, migrations.TableOauthRefreshToken)
	query = loop.FormatQuery(query)
	logging.L(s.ctx).Info("query", querySQL)

	var accessID string

	err := s.db.QueryRow(
		s.ctx,
		query,
		aT.ID,
		aT.UserId,
		aT.ClientId,
		aT.Name,
		aT.Scopes,
		aT.Revoked,
		aT.CreatedAt,
		aT.UpdatedAt,
		aT.ExpiresAt,
		rT.ID,
		rT.Revoked,
		rT.ExpiresAt,
	).Scan(&accessID)

	if err != nil {
		logging.L(s.ctx).Error("error query", err)
		return err
	}

	return nil
}

func (s *Storage) ExistsToken(aT *accessToken.AccessToken) (bool, error) {
	const op = "storage.pgsql.oauth.access-token.ExistsToken"
	logging.L(s.ctx).Info("op", op)
	var isExists bool
	querySQL := `
		SELECT (COUNT(*) > 0) as isExists
		FROM %s
		WHERE id = $1 AND user_id = $2 AND client_id = $3
	`
	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthAccessToken)
	querySQL = loop.FormatQuery(querySQL)
	logging.L(s.ctx).Info("query", querySQL)

	err := s.db.QueryRow(
		s.ctx,
		querySQL,
		aT.ID,
		aT.UserId,
		aT.ClientId,
	).Scan(&isExists)

	if err != nil {
		logging.L(s.ctx).Error("error query", err)
		return false, err
	}

	return isExists, nil
}

func (s *Storage) UpdateToken(aT *accessToken.AccessToken) (bool, error) {
	const op = "storage.pgsql.oauth.access-token.UpdateToken"
	logging.L(s.ctx).Info("op", op)

	querySQL := `
		UPDATE %s
		SET revoked = true
		WHERE id = $1 AND user_id = $2 AND client_id = $3
	`
	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthAccessToken)
	querySQL = loop.FormatQuery(querySQL)
	logging.L(s.ctx).Info("query", querySQL)
	_, err := s.db.Exec(
		s.ctx,
		querySQL,
		aT.ID,
		aT.UserId,
		aT.ClientId,
	)

	if err != nil {
		logging.L(s.ctx).Error("error query", err)
		return false, err
	}

	return true, nil
}
