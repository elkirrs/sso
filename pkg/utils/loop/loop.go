package loop

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrAttemptsExhausted = errors.New("all attempts exhausted")

func DoWithAttempt(ctx context.Context, fn func() error, attempts int, delay time.Duration) error {
	if attempts <= 0 {
		return errors.New("attempts must be greater than zero")
	}

	if delay < 0 {
		return errors.New("delay must be non-negative")
	}

	var lastErr error

	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if i < attempts-1 && delay > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	if lastErr != nil {
		return lastErr
	}
	return ErrAttemptsExhausted
}

func FormatQuery(q string) string {
	return strings.ReplaceAll(strings.ReplaceAll(q, "\t", ""), "\n", " ")
}
