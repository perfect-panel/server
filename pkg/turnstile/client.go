package turnstile

import (
	"context"
	"time"
)

// Config is the configuration for the service.
type Config struct {
	// Secret is the secret key used to verify the token.
	// This is required.
	Secret string

	// Timeout is the timeout for the service.
	// This is optional.
	// Default: 10 seconds
	Timeout time.Duration
}

// Service is the interface for the service.
// It is used to verify the token.
// It is also used to generate a random UUID.
type Service interface {
	// Verify is used to verify the token.
	// It returns true if the token is valid.
	// It returns false if the token is invalid.
	// It returns an error if there was an error verifying the token.
	Verify(ctx context.Context, token string, ip string) (bool, error)

	// VerifyIdempotent is used to verify the token.
	// The key parameter is used to ensure idempotency.
	// You may use the RandomUUID method to generate a random UUID.
	// It returns true if the token is valid.
	// It returns false if the token is invalid.
	// It returns an error if there was an error verifying the token.
	VerifyIdempotent(ctx context.Context, token string, ip string, key string) (bool, error)

	// RandomUUID is used to generate a random UUID.
	// It returns a random UUID.
	RandomUUID() string
}

// New is used to create a new service.
// It returns a new service.
func New(config Config) Service {
	return newService(config)
}
