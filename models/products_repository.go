package models

import (
	"context"

	"github.com/shopspring/decimal"
)

type GetAllProductsRequest struct {
	Offset   int
	Limit    int
	Category string
	MaxPrice *decimal.Decimal
}

type ProductsRepository interface {
	GetAllProducts(ctx context.Context, req GetAllProductsRequest) (products []Product, total int64, err error)
	GetProductByCode(ctx context.Context, code string) (*Product, error)
}
