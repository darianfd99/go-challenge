package postgres

import (
	"github.com/mytheresa/go-hiring-challenge/models"
	"gorm.io/gorm"
)

type CategoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) *CategoriesRepository {
	return &CategoriesRepository{
		db: db,
	}
}

func (r *CategoriesRepository) GetAllCategories() ([]models.Category, error) {
	var categories []models.Category

	if err := r.db.Order("id").Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}
