package categories

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
)

type fakeCategoriesRepository struct {
	categories []models.Category
	err        error
}

func (f *fakeCategoriesRepository) GetAllCategories() ([]models.Category, error) {
	return f.categories, f.err
}

func TestHandleGet(t *testing.T) {
	t.Run("returns all categories", func(t *testing.T) {
		repo := &fakeCategoriesRepository{
			categories: []models.Category{
				{Code: "clothing", Name: "Clothing"},
				{Code: "shoes", Name: "Shoes"},
				{Code: "accessories", Name: "Accessories"},
			},
		}
		h := NewCategoriesHandler(repo)

		r := httptest.NewRequest(http.MethodGet, "/categories", nil)
		w := httptest.NewRecorder()

		h.HandleGet(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		expected := `{
			"categories": [
				{"code": "clothing", "name": "Clothing"},
				{"code": "shoes", "name": "Shoes"},
				{"code": "accessories", "name": "Accessories"}
			]
		}`
		assert.JSONEq(t, expected, w.Body.String())
	})

	t.Run("returns an empty list when there are no categories", func(t *testing.T) {
		repo := &fakeCategoriesRepository{categories: []models.Category{}}
		h := NewCategoriesHandler(repo)

		r := httptest.NewRequest(http.MethodGet, "/categories", nil)
		w := httptest.NewRecorder()

		h.HandleGet(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"categories": []}`, w.Body.String())
	})

	t.Run("returns 500 when the repository fails", func(t *testing.T) {
		repo := &fakeCategoriesRepository{err: errors.New("db error")}
		h := NewCategoriesHandler(repo)

		r := httptest.NewRequest(http.MethodGet, "/categories", nil)
		w := httptest.NewRecorder()

		h.HandleGet(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "db error", "the raw repository error must not leak to the client")
		assert.JSONEq(t, `{"error": "internal server error"}`, w.Body.String())
	})
}
