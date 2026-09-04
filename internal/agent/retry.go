package agent

import (
	"context"
	"errors"
	"net/http"
)

func isRetriableError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var target *retriableError

	return errors.As(err, &target)
}

func isRetriableStatus(status int) bool {
	return status >= http.StatusInternalServerError || status == http.StatusTooManyRequests
}
