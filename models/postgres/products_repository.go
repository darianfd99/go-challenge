package postgres

import (
	"database/sql"

	"github.com/mytheresa/go-hiring-challenge/models"
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

func (r *ProductsRepository) GetAllProducts(req models.GetAllProductsRequest) ([]models.Product, int64, error) {
	var (
		total    int64
		products []models.Product
	)

	// REPEATABLE READ pins Count and Find to the same snapshot, so total can't
	// drift out of sync with the page if rows are inserted/deleted concurrently.
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Product{}).Count(&total).Error; err != nil {
			return err
		}
		return tx.Preload("Variants").Preload("Category").
			Offset(req.Offset).Limit(req.Limit).
			Find(&products).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
