package pgsql

import (
	"app/internal/config"
	"app/pkg/common/logging"
	"app/pkg/utils/loop"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(
	ctx context.Context,
	cfg config.DB,
) (pool *pgxpool.Pool, err error) {

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.PGSQL.Username,
		cfg.PGSQL.Password,
		cfg.PGSQL.Host,
		cfg.PGSQL.Port,
		cfg.PGSQL.Database,
		cfg.PGSQL.SSLMode,
	)

	pgxCfg, parseConfigErr := pgxpool.ParseConfig(dsn)
	if parseConfigErr != nil {
		logging.L(ctx).Error("Unable to parse config: ", fmt.Sprintf("%v\n", parseConfigErr))
		return nil, parseConfigErr
	}

	pool, parseConfigErr = pgxpool.NewWithConfig(ctx, pgxCfg)
	if parseConfigErr != nil {
		logging.L(ctx).Error("Failed to parse Postgres SQL configuration due to error", fmt.Sprintf("%v\n", parseConfigErr))
		return nil, parseConfigErr
	}

	err = loop.DoWithAttempt(func() error {
		pingErr := pool.Ping(ctx)
		if pingErr != nil {
			logging.L(ctx).Warn("Failed to connect to postgres due to error", pingErr)
			logging.L(ctx).Warn("Going to do the next attempt")
			return pingErr
		}

		return nil
	}, cfg.MaxAttempts, cfg.MaxDelay)

	if err != nil {
		logging.L(ctx).Error("All attempts are exceeded. Unable to connect to Postgres SQL")
		return nil, err
	}

	return pool, nil
}
