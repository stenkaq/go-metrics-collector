package repository

import (
	"context"
	"errors"
	"net"
	"syscall"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const maxRetries = 3

func isRetriable(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgerrcode.IsConnectionException(pgErr.Code)
	}

	if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) {
		return true
	}

	var netErr net.Error

	return errors.As(err, &netErr)
}

func withRetry(ctx context.Context, f func() error) error {
	err := f()
	delay := 1 * time.Second

	for range maxRetries {
		if !isRetriable(err) {
			return err
		}

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}

		err = f()
		delay += 2 * time.Second
	}

	return err
}
