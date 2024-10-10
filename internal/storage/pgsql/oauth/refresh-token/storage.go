package refresh_token

import (
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

func (s *Storage) CreateRefreshToken(rT *refreshTokenDomain.RefreshToken) (string, error) {
	const op = "storage.pgsql.oauth.refresh-token.CreateRefreshToken"
	logging.L(s.ctx).Info("op", op)

	querySQL := `
			INSERT INTO %s (id, access_token_id, revoked, expires_at)
			VALUES ($1, $2, $3, $4)
			`
	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthRefreshToken)
	querySQL = loop.FormatQuery(querySQL)
	logging.L(s.ctx).Info("query", querySQL)

	_, err := s.db.Exec(
		s.ctx,
		querySQL,
		rT.ID,
		rT.AccessTokenId,
		rT.Revoked,
		rT.ExpiresAt,
	)

	if err != nil {
		logging.L(s.ctx).Error("error query", err)
		return "", err
	}

	return rT.ID, nil
}

func (s *Storage) ExistsToken(rT *refreshTokenDomain.RefreshToken) (bool, error) {
	const op = "storage.pgsql.oauth.refresh-token.ExistsToken"
	logging.L(s.ctx).Info("op", op)
	var isExists bool
	querySQL := `
		SELECT (COUNT(*) > 0) as isExists
		FROM %s
		WHERE id = $1 AND access_token_id = $2
	`
	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthRefreshToken)
	querySQL = loop.FormatQuery(querySQL)
	logging.L(s.ctx).Info("query", querySQL)

	err := s.db.QueryRow(
		s.ctx,
		querySQL,
		rT.ID,
		rT.AccessTokenId,
	).Scan(&isExists)

	if err != nil {
		logging.L(s.ctx).Error("error query", err)
		return false, err
	}

	return isExists, nil
}

func (s *Storage) GetToken(rT *refreshTokenDomain.RefreshToken) (refreshTokenDomain.RefreshToken, error) {
	const op = "storage.pgsql.oauth.refresh-token.ExistsToken"
	logging.L(s.ctx).Info("op", op)
	var rTQ = refreshTokenDomain.RefreshToken{}

	querySQL := `
		SELECT id, access_token_id, revoked, expires_at
		FROM %s
		WHERE id = $1 AND access_token_id = $2
		ORDER BY expires_at DESC
	`
	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthRefreshToken)
	querySQL = loop.FormatQuery(querySQL)
	logging.L(s.ctx).Info("query", querySQL)

	err := s.db.QueryRow(
		s.ctx,
		querySQL,
		rT.ID,
		rT.AccessTokenId,
	).Scan(
		&rTQ.ID,
		&rTQ.AccessTokenId,
		&rTQ.Revoked,
		&rTQ.ExpiresAt,
	)

	if err != nil {
		logging.L(s.ctx).Error("error query", err)
		return refreshTokenDomain.RefreshToken{}, err
	}

	return rTQ, nil
}

func (s *Storage) UpdateToken(rT *refreshTokenDomain.RefreshToken) (bool, error) {
	const op = "storage.pgsql.oauth.refresh-token.UpdateToken"
	logging.L(s.ctx).Info("op", op)

	querySQL := `
		UPDATE %s
		SET revoked = true
		WHERE id = $1 AND access_token_id = $2
	`
	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthRefreshToken)
	querySQL = loop.FormatQuery(querySQL)
	logging.L(s.ctx).Info("query", querySQL)
	_, err := s.db.Exec(
		s.ctx,
		querySQL,
		rT.ID,
		rT.AccessTokenId,
	)

	if err != nil {
		logging.L(s.ctx).Error("error query", err)
		return false, err
	}

	return true, nil
}

func (s *Storage) GetLastReceivedToken(rT *refreshTokenDomain.RefreshToken) (refreshTokenDomain.RefreshToken, error) {
	const op = "storage.pgsql.oauth.refresh-token.GetLastReceivedToken"
	logging.L(s.ctx).Info("op", op)
	var rTQ = refreshTokenDomain.RefreshToken{}

	querySQL := `
		SELECT oauth_refresh_tokens.id,
			   oauth_refresh_tokens.access_token_id,
			   oauth_refresh_tokens.revoked,
			   oauth_refresh_tokens.expires_at
		FROM oauth_access_tokens
		INNER JOIN oauth_refresh_tokens on oauth_access_tokens.id = oauth_refresh_tokens.access_token_id
		WHERE oauth_access_tokens.user_id = (SELECT oauth_access_tokens.user_id
							FROM oauth_refresh_tokens
							INNER JOIN oauth_access_tokens
								ON oauth_access_tokens.id = oauth_refresh_tokens.access_token_id
							WHERE oauth_refresh_tokens.id = $1
								AND oauth_refresh_tokens.access_token_id = $2)
		ORDER BY oauth_refresh_tokens.expires_at DESC
		LIMIT 1
	`
	querySQL = loop.FormatQuery(querySQL)
	logging.L(s.ctx).Info("query", querySQL)

	err := s.db.QueryRow(
		s.ctx,
		querySQL,
		rT.ID,
		rT.AccessTokenId,
	).Scan(
		&rTQ.ID,
		&rTQ.AccessTokenId,
		&rTQ.Revoked,
		&rTQ.ExpiresAt,
	)

	if err != nil {
		logging.L(s.ctx).Error("error query", err)
		return refreshTokenDomain.RefreshToken{}, err
	}

	return rTQ, nil
}
