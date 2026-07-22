package catalog

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
)

const (
	defaultOffset = 0
	defaultLimit  = 10
	minLimit      = 1
	maxLimit      = 100
)

type Response struct {
	Products []Product `json:"products"`
	Total    int64     `json:"total"`
}

type Product struct {
	Code     string       `json:"code"`
	Price    float64      `json:"price"`
	Category api.Category `json:"category"`
}

type ProductDetails struct {
	Code     string          `json:"code"`
	Price    float64         `json:"price"`
	Category api.Category    `json:"category"`
	Variants []VariantDetail `json:"variants"`
}

type VariantDetail struct {
	SKU   string  `json:"sku"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type CatalogHandler struct {
	repo models.ProductsRepository
}

func NewCatalogHandler(r models.ProductsRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

func newCategory(c models.Category) api.Category {
	return api.Category{Code: c.Code, Name: c.Name}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	req, err := parseGetAllProductsRequest(r)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	res, total, err := h.repo.GetAllProducts(req)
	if err != nil {
		api.InternalError(w, "catalog", err)
		return
	}

	// Map response
	products := make([]Product, len(res))
	for i, p := range res {
		products[i] = Product{
			Code:     p.Code,
			Price:    p.Price.InexactFloat64(),
			Category: newCategory(p.Category),
		}
	}

	api.OKResponse(w, Response{
		Products: products,
		Total:    total,
	})
}

func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	p, err := h.repo.GetProductByCode(code)
	if errors.Is(err, models.ErrProductNotFound) {
		api.ErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		api.InternalError(w, "catalog", err)
		return
	}

	variants := make([]VariantDetail, len(p.Variants))
	for i, v := range p.Variants {
		price := p.Price
		if v.Price != nil {
			price = *v.Price
		}
		variants[i] = VariantDetail{
			SKU:   v.SKU,
			Name:  v.Name,
			Price: price.InexactFloat64(),
		}
	}

	api.OKResponse(w, ProductDetails{
		Code:     p.Code,
		Price:    p.Price.InexactFloat64(),
		Category: newCategory(p.Category),
		Variants: variants,
	})
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
	offset = defaultOffset
	if v := r.URL.Query().Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, 0, ErrInvalidOffset
		}
		offset = n
	}
	if offset < 0 {
		offset = defaultOffset
	}

	limit = defaultLimit
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, 0, ErrInvalidLimit
		}
		limit = n
	}
	limit = min(max(limit, minLimit), maxLimit)

	return offset, limit, nil
}
