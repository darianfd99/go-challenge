package models

type ProductsRepository interface {
	GetAllProducts() ([]Product, error)
}
