package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

func byCategory(code string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if code == "" {
			return db
		}
		// Joins("Category") aliases the joined table as the quoted, case-preserved
		// "Category"; an unquoted reference here would fold to lowercase in
		// Postgres and no longer match the alias.
		return db.Where(`"Category".code = ?`, code)
	}
}

func priceLessThan(max *decimal.Decimal) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if max == nil {
			return db
		}
		return db.Where("price < ?", max)
	}
}

func (r *ProductsRepository) GetAllProducts(ctx context.Context, req models.GetAllProductsRequest) ([]models.Product, int64, error) {
	var (
		total    int64
		products []models.Product
	)

	// REPEATABLE READ pins Count and Find to the same snapshot, so total can't
	// drift out of sync with the page if rows are inserted/deleted concurrently.
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&models.Product{}).
			Joins("Category").
			Scopes(byCategory(req.Category), priceLessThan(req.MaxPrice))

		if err := query.Count(&total).Error; err != nil {
			return err
		}
		return query.Order("products.id").
			Offset(req.Offset).Limit(req.Limit).
			Find(&products).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductsRepository) GetProductByCode(ctx context.Context, code string) (*models.Product, error) {
	var product models.Product

	err := r.db.WithContext(ctx).Joins("Category").
		Preload("Variants").
		Where("products.code = ?", code).
		First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, models.ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}

	return &product, nil
}
