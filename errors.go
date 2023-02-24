package replicateapi

import (
	"net/http"

	"github.com/pkg/errors"
)

var (
	// ErrUnauthorized you should check your authorization token and availability of the model
	ErrUnauthorized = errors.New("unauthorized")
	// ErrRateLimitReached check the official docs regarding the current limits https://replicate.com/docs/reference/http#rate-limits
	ErrRateLimitReached = errors.New("rate limit reached")
)

func handleStatusCode(code int) error {
	switch code {
	case http.StatusTooManyRequests:
		return ErrRateLimitReached
	case http.StatusUnauthorized:
		return ErrUnauthorized
	}
	return nil
}
