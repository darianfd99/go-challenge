package models

import "context"

type CategoriesRepository interface {
	GetAllCategories(ctx context.Context) ([]Category, error)
	// CreateCategory returns a non-nil Category whenever err is nil.
	CreateCategory(ctx context.Context, category Category) (*Category, error)
}
