package models

import "github.com/shopspring/decimal"

type GetAllProductsRequest struct {
	Offset   int
	Limit    int
	Category string
	MaxPrice *decimal.Decimal
}

type ProductsRepository interface {
	GetAllProducts(req GetAllProductsRequest) (products []Product, total int64, err error)
	GetProductByCode(code string) (*Product, error)
}
