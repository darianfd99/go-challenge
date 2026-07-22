package catalog

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
)

type Response struct {
	Products []Product `json:"products"`
	Total    int64     `json:"total"`
}

type Product struct {
	Code     string   `json:"code"`
	Price    float64  `json:"price"`
	Category Category `json:"category"`
}

type Category struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CatalogHandler struct {
	repo models.ProductsRepository
}

func NewCatalogHandler(r models.ProductsRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	req, err := parseGetAllProductsRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, total, err := h.repo.GetAllProducts(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Map response
	products := make([]Product, len(res))
	for i, p := range res {
		products[i] = Product{
			Code:  p.Code,
			Price: p.Price.InexactFloat64(),
			Category: Category{
				Code: p.Category.Code,
				Name: p.Category.Name,
			},
		}
	}

	// Return the products as a JSON response
	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Products: products,
		Total:    total,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func parseGetAllProductsRequest(r *http.Request) (models.GetAllProductsRequest, error) {
	offset, limit, err := parsePagination(r)
	if err != nil {
		return models.GetAllProductsRequest{}, err
	}

	category, maxPrice, err := parseFilters(r)
	if err != nil {
		return models.GetAllProductsRequest{}, err
	}

	return models.GetAllProductsRequest{
		Offset:   offset,
		Limit:    limit,
		Category: category,
		MaxPrice: maxPrice,
	}, nil
}

func parseFilters(r *http.Request) (category string, maxPrice *decimal.Decimal, err error) {
	category = r.URL.Query().Get("category")

	if v := r.URL.Query().Get("price_lt"); v != "" {
		p, err := decimal.NewFromString(v)
		if err != nil {
			return "", nil, ErrInvalidPrice
		}
		if p.IsNegative() {
			return "", nil, ErrInvalidPrice
		}
		maxPrice = &p
	}

	return category, maxPrice, nil
}

func parsePagination(r *http.Request) (offset int, limit int, err error) {
	offset = 0
	if v := r.URL.Query().Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, 0, ErrInvalidOffset
		}
		offset = n
	}
	if offset < 0 {
		offset = 0
	}

	limit = 10
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, 0, ErrInvalidLimit
		}
		limit = n
	}
	limit = min(max(limit, 1), 100)

	return offset, limit, nil
}
