package catalog

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type fakeProductsRepository struct {
	product *models.Product
	err     error
}

func (f *fakeProductsRepository) GetAllProducts(req models.GetAllProductsRequest) ([]models.Product, int64, error) {
	return nil, 0, nil
}

func (f *fakeProductsRepository) GetProductByCode(code string) (*models.Product, error) {
	return f.product, f.err
}

func TestHandleGetByCode(t *testing.T) {
	productPrice := decimal.NewFromFloat(100.00)
	variantPrice := decimal.NewFromFloat(120.00)

	t.Run("returns product with variants, inheriting product price when variant price is nil", func(t *testing.T) {
		repo := &fakeProductsRepository{
			product: &models.Product{
				Code:  "PROD001",
				Price: productPrice,
				Category: models.Category{
					Code: "clothing",
					Name: "Clothing",
				},
				Variants: []models.Variant{
					{SKU: "SKU001", Name: "Small", Price: &variantPrice},
					{SKU: "SKU002", Name: "Large", Price: nil},
				},
			},
		}
		h := NewCatalogHandler(repo)

		r := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
		r.SetPathValue("code", "PROD001")
		w := httptest.NewRecorder()

		h.HandleGetByCode(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		expected := `{
			"code": "PROD001",
			"price": 100,
			"category": {"code": "clothing", "name": "Clothing"},
			"variants": [
				{"sku": "SKU001", "name": "Small", "price": 120},
				{"sku": "SKU002", "name": "Large", "price": 100}
			]
		}`
		assert.JSONEq(t, expected, w.Body.String())
	})

	t.Run("returns 404 when product is not found", func(t *testing.T) {
		repo := &fakeProductsRepository{err: models.ErrProductNotFound}
		h := NewCatalogHandler(repo)

		r := httptest.NewRequest(http.MethodGet, "/catalog/UNKNOWN", nil)
		r.SetPathValue("code", "UNKNOWN")
		w := httptest.NewRecorder()

		h.HandleGetByCode(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, `{"error": "product not found"}`, w.Body.String())
	})

	t.Run("returns 500 when the repository fails", func(t *testing.T) {
		repo := &fakeProductsRepository{err: errors.New("db error")}
		h := NewCatalogHandler(repo)

		r := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
		r.SetPathValue("code", "PROD001")
		w := httptest.NewRecorder()

		h.HandleGetByCode(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "db error", "the raw repository error must not leak to the client")
		assert.JSONEq(t, `{"error": "internal server error"}`, w.Body.String())
	})
}
