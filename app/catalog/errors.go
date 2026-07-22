package catalog

import "errors"

var (
	ErrInvalidOffset = errors.New("invalid offset")
	ErrInvalidLimit  = errors.New("invalid limit")
	ErrInvalidPrice  = errors.New("invalid price_lt")

	// errInternal is the message returned to clients for any 500; the real
	// error is logged server-side instead, since it may contain details
	// (e.g. raw DB error text) that shouldn't be exposed externally.
	errInternal = errors.New("internal server error")
)
