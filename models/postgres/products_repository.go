package postgres

import (
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

func (r *ProductsRepository) GetAllProducts() ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Preload("Variants").Preload("Category").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}
