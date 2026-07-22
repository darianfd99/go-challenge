package categories

import "errors"

var (
	ErrInvalidRequestBody = errors.New("invalid request body")
	ErrMissingFields      = errors.New("code and name are required")
)
