package http

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
	"fintrack-go/internal/db"
	"fintrack-go/internal/validator"
)

type CategoryHandler struct {
	*Handler
	db db.Database
}

func NewCategoryHandler(logger zerolog.Logger, database db.Database) *CategoryHandler {
	return &CategoryHandler{
		Handler: NewHandler(logger),
		db:      database,
	}
}

type CreateCategoryRequest struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := validator.ValidateUUID(req.UserID); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error(), map[string]string{
			"field": "user_id",
			"value": req.UserID,
		})
		return
	}

	if err := validator.ValidateCategoryName(req.Name); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error(), map[string]string{
			"field": "name",
			"value": req.Name,
		})
		return
	}

	category, err := h.db.CreateCategory(r.Context(), req.UserID, req.Name)
	if err != nil {
		if err == db.ErrUserNotFound {
			h.respondWithError(w, http.StatusNotFound, "User not found", nil)
			return
		}
		if err == db.ErrDuplicateCategory {
			h.respondWithError(w, http.StatusConflict, "Category name already exists for this user", nil)
			return
		}
		h.Logger.Error().Err(err).Msg("Failed to create category")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create category", nil)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, category)
}

func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.respondWithError(w, http.StatusBadRequest, "user_id query parameter is required", nil)
		return
	}

	if err := validator.ValidateUUID(userID); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error(), map[string]string{
			"field": "user_id",
			"value": userID,
		})
		return
	}

	categories, err := h.db.ListCategories(r.Context(), userID)
	if err != nil {
		h.Logger.Error().Err(err).Msg("Failed to list categories")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list categories", nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, categories)
}
