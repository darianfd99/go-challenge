package catalog

import (
	"context"
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

	products []models.Product
	total    int64
	listErr  error
	gotReq   models.GetAllProductsRequest
}

func (f *fakeProductsRepository) GetAllProducts(_ context.Context, req models.GetAllProductsRequest) ([]models.Product, int64, error) {
	f.gotReq = req
	return f.products, f.total, f.listErr
}

func (f *fakeProductsRepository) GetProductByCode(_ context.Context, code string) (*models.Product, error) {
	return f.product, f.err
}

func TestHandleGet(t *testing.T) {
	t.Run("returns products with category and total", func(t *testing.T) {
		repo := &fakeProductsRepository{
			products: []models.Product{
				{
					Code:  "PROD001",
					Price: decimal.NewFromFloat(10.99),
					Category: models.Category{
						Code: "clothing",
						Name: "Clothing",
					},
				},
				{
					Code:  "PROD002",
					Price: decimal.NewFromFloat(12.49),
					Category: models.Category{
						Code: "shoes",
						Name: "Shoes",
					},
				},
			},
			total: 8,
		}
		h := NewCatalogHandler(repo)

		r := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		w := httptest.NewRecorder()

		h.HandleGet(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		expected := `{
			"products": [
				{"code": "PROD001", "price": 10.99, "category": {"code": "clothing", "name": "Clothing"}},
				{"code": "PROD002", "price": 12.49, "category": {"code": "shoes", "name": "Shoes"}}
			],
			"total": 8
		}`
		assert.JSONEq(t, expected, w.Body.String())
	})

	t.Run("returns an empty list when there are no products", func(t *testing.T) {
		repo := &fakeProductsRepository{products: []models.Product{}}
		h := NewCatalogHandler(repo)

		r := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		w := httptest.NewRecorder()

		h.HandleGet(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"products": [], "total": 0}`, w.Body.String())
	})

	t.Run("returns 500 when the repository fails", func(t *testing.T) {
		repo := &fakeProductsRepository{listErr: errors.New("db error")}
		h := NewCatalogHandler(repo)

		r := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		w := httptest.NewRecorder()

		h.HandleGet(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "db error", "the raw repository error must not leak to the client")
		assert.JSONEq(t, `{"error": "internal server error"}`, w.Body.String())
	})

	t.Run("pagination and filters", func(t *testing.T) {
		price := decimal.NewFromFloat(19.99)

		cases := []struct {
			name         string
			query        string
			wantStatus   int
			wantErrBody  string
			wantOffset   int
			wantLimit    int
			wantCategory string
			wantMaxPrice *decimal.Decimal
		}{
			{
				name:       "defaults when no params given",
				query:      "",
				wantStatus: http.StatusOK,
				wantOffset: 0,
				wantLimit:  10,
			},
			{
				name:       "custom offset and limit",
				query:      "?offset=5&limit=20",
				wantStatus: http.StatusOK,
				wantOffset: 5,
				wantLimit:  20,
			},
			{
				name:       "negative offset clamped to 0",
				query:      "?offset=-5",
				wantStatus: http.StatusOK,
				wantOffset: 0,
				wantLimit:  10,
			},
			{
				name:       "limit below minimum clamped to 1",
				query:      "?limit=0",
				wantStatus: http.StatusOK,
				wantOffset: 0,
				wantLimit:  1,
			},
			{
				name:       "limit above maximum clamped to 100",
				query:      "?limit=500",
				wantStatus: http.StatusOK,
				wantOffset: 0,
				wantLimit:  100,
			},
			{
				name:        "non-numeric offset returns 400",
				query:       "?offset=abc",
				wantStatus:  http.StatusBadRequest,
				wantErrBody: `{"error": "invalid offset"}`,
			},
			{
				name:        "non-numeric limit returns 400",
				query:       "?limit=abc",
				wantStatus:  http.StatusBadRequest,
				wantErrBody: `{"error": "invalid limit"}`,
			},
			{
				name:         "category filter passed through",
				query:        "?category=clothing",
				wantStatus:   http.StatusOK,
				wantOffset:   0,
				wantLimit:    10,
				wantCategory: "clothing",
			},
			{
				name:         "price_lt filter passed through",
				query:        "?price_lt=19.99",
				wantStatus:   http.StatusOK,
				wantOffset:   0,
				wantLimit:    10,
				wantMaxPrice: &price,
			},
			{
				name:        "negative price_lt returns 400",
				query:       "?price_lt=-5",
				wantStatus:  http.StatusBadRequest,
				wantErrBody: `{"error": "invalid price_lt"}`,
			},
			{
				name:        "non-numeric price_lt returns 400",
				query:       "?price_lt=abc",
				wantStatus:  http.StatusBadRequest,
				wantErrBody: `{"error": "invalid price_lt"}`,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				repo := &fakeProductsRepository{products: []models.Product{}}
				h := NewCatalogHandler(repo)

				r := httptest.NewRequest(http.MethodGet, "/catalog"+tc.query, nil)
				w := httptest.NewRecorder()

				h.HandleGet(w, r)

				assert.Equal(t, tc.wantStatus, w.Code)

				if tc.wantStatus != http.StatusOK {
					assert.JSONEq(t, tc.wantErrBody, w.Body.String())
					return
				}

				assert.Equal(t, tc.wantOffset, repo.gotReq.Offset)
				assert.Equal(t, tc.wantLimit, repo.gotReq.Limit)
				assert.Equal(t, tc.wantCategory, repo.gotReq.Category)

				if tc.wantMaxPrice == nil {
					assert.Nil(t, repo.gotReq.MaxPrice)
				} else if assert.NotNil(t, repo.gotReq.MaxPrice) {
					assert.True(t, tc.wantMaxPrice.Equal(*repo.gotReq.MaxPrice))
				}
			})
		}
	})
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
