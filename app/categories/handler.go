package categories

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type Response struct {
	Categories []Category `json:"categories"`
}

type Category struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CreateCategoryRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CategoriesHandler struct {
	repo models.CategoriesRepository
}

func NewCategoriesHandler(r models.CategoriesRepository) *CategoriesHandler {
	return &CategoriesHandler{
		repo: r,
	}
}

func (h *CategoriesHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	cats, err := h.repo.GetAllCategories()
	if err != nil {
		log.Printf("categories: %s", err)
		api.ErrorResponse(w, http.StatusInternalServerError, api.ErrInternalServerError.Error())
		return
	}

	categories := make([]Category, len(cats))
	for i, c := range cats {
		categories[i] = Category{
			Code: c.Code,
			Name: c.Name,
		}
	}

	api.OKResponse(w, Response{Categories: categories})
}

func (h *CategoriesHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, ErrInvalidRequestBody.Error())
		return
	}
	if req.Code == "" || req.Name == "" {
		api.ErrorResponse(w, http.StatusBadRequest, ErrMissingFields.Error())
		return
	}

	created, err := h.repo.CreateCategory(models.Category{Code: req.Code, Name: req.Name})
	if errors.Is(err, models.ErrCategoryCodeExists) {
		api.ErrorResponse(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		log.Printf("categories: %s", err)
		api.ErrorResponse(w, http.StatusInternalServerError, api.ErrInternalServerError.Error())
		return
	}

	api.CreatedResponse(w, Category{Code: created.Code, Name: created.Name})
}
