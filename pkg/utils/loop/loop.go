package loop

import (
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"strings"
	"time"
)

func DoWithAttempt(fn func() error, attempts int, delay time.Duration) (err error) {
	for attempts > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			attempts--
			continue
		}
		return nil
	}
	return
}

func FormatQuery(q string) string {
	return strings.ReplaceAll(strings.ReplaceAll(q, "\t", ""), "\n", " ")
}

func ParsePgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.Is(err, pgErr) {
		pgErr = err.(*pgconn.PgError)
		return fmt.Errorf("database error. message:%s, detail:%s, where:%s, sqlstate:%s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.SQLState())
	}
	return err
}
