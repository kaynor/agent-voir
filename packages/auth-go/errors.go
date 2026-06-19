package auth

import "errors"

var (
	// ErrMissingBearer is returned when Authorization header is absent or malformed.
	ErrMissingBearer = errors.New("missing bearer token")
	// ErrUnauthorized is returned when credentials are invalid.
	ErrUnauthorized = errors.New("unauthorized")
)
