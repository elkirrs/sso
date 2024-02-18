package user

import (
	"app/internal/domain/user"
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

func New(pgClient *pgxpool.Pool, log *slog.Logger, ctx context.Context) (*Storage, error) {
	return &Storage{
		db:  pgClient,
		log: log,
		ctx: ctx,
	}, nil
}

func (s *Storage) Registration(req *user.CreateUser) (error error) {
	const op = "storage.pgsql.user.Registration"

	querySQL := `INSERT INTO users (uuid, name, email, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	querySQL = loop.FormatQuery(querySQL)

	s.log.With(
		slog.String("op", op),
		slog.String("sql query", querySQL),
	).Info("prepared query")

	_, err := s.db.Exec(
		s.ctx,
		querySQL,
		req.UUID,
		req.Name,
		req.Email,
		req.Password,
		req.CreatedAt,
		req.UpdatedAt,
	)

	if err != nil {
		s.log.Error("error query", err)
		return err
	}

	return nil
}

func (s *Storage) Login(req *user.User) (user.User, error) {
	const op = "storage.pgsql.user.login"

	querySQL := `SELECT id, uuid, name, email, password FROM %s WHERE name = $1 OR email = $2`
	querySQL = fmt.Sprintf(querySQL, migrations.TableUsers)
	querySQL = loop.FormatQuery(querySQL)

	s.log.With(
		slog.String("op", op),
		slog.String("sql query", querySQL),
	).Info("prepared query")

	var usrStorage user.User

	err := s.db.QueryRow(
		s.ctx,
		querySQL,
		req.Name,
		req.Email,
	).Scan(
		&usrStorage.ID,
		&usrStorage.UUID,
		&usrStorage.Name,
		&usrStorage.Email,
		&usrStorage.Password,
	)

	if err != nil {
		s.log.Error("error query", err)
		return usrStorage, err
	}

	return usrStorage, nil
}
