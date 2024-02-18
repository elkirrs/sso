package main

import (
	"app/internal/config"
	"app/migrations"
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"log/slog"
	"os"
)

const (
	MigrationUp   = "up"
	MigrationDown = "down"
)

func main() {
	var migrationStatus string

	fmt.Print("Upgrade or Downgrade Database? (Select up or down): ")
	fmt.Scan(&migrationStatus)

	cfg := config.MustLoad()
	log := setupLogger()
	log.Info("starting logger")

	pgDsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.PGSQL.Username,
		cfg.DB.PGSQL.Password,
		cfg.DB.PGSQL.Host,
		cfg.DB.PGSQL.Port,
		cfg.DB.PGSQL.Database,
		cfg.DB.PGSQL.SSLMode,
	)

	stdlib.GetDefaultDriver()

	log.Info("Postgres SQL Migrate initializing")

	db, err := goose.OpenDBWithDriver("postgres", pgDsn)
	if err != nil {
		panic(err)
	}
	goose.SetBaseFS(&migrations.Content)

	err = goose.SetDialect("postgres")
	if err != nil {
		panic(err)
	}

	switch migrationStatus {
	case MigrationUp:
		err = goose.Up(db, ".")
		log.Info("migration up till last")

	case MigrationDown:
		err = goose.Down(db, ".")
		log.Info("migration down till last")
	}

	if err != nil {
		panic(err)
	}

	err = db.Close()
	if err != nil {
		panic(err)
	}

	fmt.Println("migrations applied")
}

func setupLogger() *slog.Logger {
	var log *slog.Logger
	log = slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	return log
}
