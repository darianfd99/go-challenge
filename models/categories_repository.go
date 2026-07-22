package models

type CategoriesRepository interface {
	GetAllCategories() ([]Category, error)
	// CreateCategory returns a non-nil Category whenever err is nil.
	CreateCategory(category Category) (*Category, error)
}
