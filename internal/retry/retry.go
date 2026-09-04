package retry

import (
	"context"
	"time"
)

const maxAttempts = 3

func Do(ctx context.Context, isRetriable func(error) bool, f func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	err := f()
	delay := 1 * time.Second

	for range maxAttempts {
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
