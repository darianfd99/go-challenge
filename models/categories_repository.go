package models

type CategoriesRepository interface {
	GetAllCategories() ([]Category, error)
}
