package models

import "errors"

var (
	ErrProductNotFound    = errors.New("product not found")
	ErrCategoryCodeExists = errors.New("category code already exists")
)
