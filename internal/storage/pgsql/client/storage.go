package client

import (
	"app/internal/domain/client"
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

func (s *Storage) GetClient(ID string) (client.Client, error) {
	const op = "storage.pgsql.oauth.client.GetClient"
	logging.L(s.ctx).Info("op", op)

	querySQL := `
		SELECT id, user_id, name, secret, provider, redirect, personal_access_client, password_client, revoked
		FROM %s
		WHERE id = $1
	`
	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthClient)
	querySQL = loop.FormatQuery(querySQL)
	logging.L(s.ctx).Info("prepare query", querySQL)

	var c client.Client

	err := s.db.QueryRow(
		s.ctx,
		querySQL,
		ID,
	).Scan(
		&c.ID,
		&c.UserId,
		&c.Name,
		&c.Secret,
		&c.Provider,
		&c.Redirect,
		&c.PersonalAccessClient,
		&c.PasswordClient,
		&c.Revoked,
	)
	if err != nil {
		logging.L(s.ctx).Error("error query db", err)
		return c, err
	}

	return c, nil
}

func (s *Storage) CreateClient() error {
	panic("implement me")
}

func (s *Storage) GetClientByName(name string) (client.Client, error) {
	const op = "storage.pgsql.oauth.client.GetClientByName"
	logging.L(s.ctx).Info("op", op)

	querySQL := `
		SELECT id, user_id, name, secret, provider, redirect, personal_access_client, password_client, revoked
		FROM %s
		WHERE name = $1
	`

	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthClient)
	querySQL = loop.FormatQuery(querySQL)
	logging.L(s.ctx).Info("prepare query", querySQL)

	var c client.Client

	err := s.db.QueryRow(
		s.ctx,
		querySQL,
		name,
	).Scan(
		&c.ID,
		&c.UserId,
		&c.Name,
		&c.Secret,
		&c.Provider,
		&c.Redirect,
		&c.PersonalAccessClient,
		&c.PasswordClient,
		&c.Revoked,
	)

	if err != nil {
		logging.L(s.ctx).Error("error query db", err)
		return c, err
	}

	return c, nil
}