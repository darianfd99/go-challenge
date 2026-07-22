package postgres

import (
	"context"
	"errors"

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

func (r *CategoriesRepository) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category

	if err := r.db.WithContext(ctx).Order("id").Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *CategoriesRepository) CreateCategory(ctx context.Context, category models.Category) (*models.Category, error) {
	if err := r.db.WithContext(ctx).Create(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, models.ErrCategoryCodeExists
		}
		return nil, err
	}

	return &category, nil
}
