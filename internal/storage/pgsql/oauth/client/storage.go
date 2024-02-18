package client

import (
	"app/internal/domain/oauth/clients"
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

func (s *Storage) GetClient(ID string) (clients.Client, error) {
	const op = "storage.pgsql.oauth.client.GetClient"
	s.log.Info("op", op)

	querySQL := `
		SELECT id, user_id, name, secret, provider, redirect, personal_access_client, password_client, revoked
		FROM %s
		WHERE id = $1
	`
	querySQL = fmt.Sprintf(querySQL, migrations.TableOauthClient)
	querySQL = loop.FormatQuery(querySQL)
	s.log.Info("prepare query", querySQL)

	var client clients.Client

	err := s.db.QueryRow(
		s.ctx,
		querySQL,
		ID,
	).Scan(
		&client.ID,
		&client.UserId,
		&client.Name,
		&client.Secret,
		&client.Provider,
		&client.Redirect,
		&client.PersonalAccessClient,
		&client.PasswordClient,
		&client.Revoked,
	)
	s.log.Info("client data", client)
	if err != nil {
		s.log.Error("error query db", err)
		return client, err
	}

	return client, nil
}

func (s *Storage) CreateClient() error {
	panic("implement me")
}
