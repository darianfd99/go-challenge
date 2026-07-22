package models

type CategoriesRepository interface {
	GetAllCategories() ([]Category, error)
	CreateCategory(category Category) (*Category, error)
}
