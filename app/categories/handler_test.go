package categories

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
)

type fakeCategoriesRepository struct {
	categories []models.Category
	err        error

	createErr error
}

func (f *fakeCategoriesRepository) GetAllCategories(_ context.Context) ([]models.Category, error) {
	return f.categories, f.err
}

func (f *fakeCategoriesRepository) CreateCategory(_ context.Context, category models.Category) (*models.Category, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &category, nil
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

func TestHandleCreate(t *testing.T) {
	t.Run("creates a category", func(t *testing.T) {
		repo := &fakeCategoriesRepository{}
		h := NewCategoriesHandler(repo)

		r := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"code":"toys","name":"Toys"}`))
		w := httptest.NewRecorder()

		h.HandleCreate(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.JSONEq(t, `{"code":"toys","name":"Toys"}`, w.Body.String())
	})

	t.Run("returns 400 for a malformed body", func(t *testing.T) {
		repo := &fakeCategoriesRepository{}
		h := NewCategoriesHandler(repo)

		r := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader([]byte(`not json`)))
		w := httptest.NewRecorder()

		h.HandleCreate(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 when required fields are missing", func(t *testing.T) {
		repo := &fakeCategoriesRepository{}
		h := NewCategoriesHandler(repo)

		r := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"code":"toys"}`))
		w := httptest.NewRecorder()

		h.HandleCreate(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 409 when the category code already exists", func(t *testing.T) {
		repo := &fakeCategoriesRepository{createErr: models.ErrCategoryCodeExists}
		h := NewCategoriesHandler(repo)

		r := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"code":"clothing","name":"Clothing"}`))
		w := httptest.NewRecorder()

		h.HandleCreate(w, r)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.JSONEq(t, `{"error": "category code already exists"}`, w.Body.String())
	})

	t.Run("returns 500 when the repository fails", func(t *testing.T) {
		repo := &fakeCategoriesRepository{createErr: errors.New("db error")}
		h := NewCategoriesHandler(repo)

		r := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"code":"toys","name":"Toys"}`))
		w := httptest.NewRecorder()

		h.HandleCreate(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "db error", "the raw repository error must not leak to the client")
		assert.JSONEq(t, `{"error": "internal server error"}`, w.Body.String())
	})
}
