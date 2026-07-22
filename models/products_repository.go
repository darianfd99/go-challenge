package models

type GetAllProductsRequest struct {
	Offset int
	Limit  int
}

type ProductsRepository interface {
	GetAllProducts(req GetAllProductsRequest) (products []Product, total int64, err error)
}
