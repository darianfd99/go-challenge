package catalog

import "errors"

var (
	ErrInvalidOffset = errors.New("invalid offset")
	ErrInvalidLimit  = errors.New("invalid limit")
	ErrInvalidPrice  = errors.New("invalid price_lt")
)
