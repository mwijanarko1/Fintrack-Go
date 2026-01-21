package http

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
	"fintrack-go/internal/db"
	"fintrack-go/internal/validator"
)

type UserHandler struct {
	*Handler
	db db.Database
}

func NewUserHandler(logger zerolog.Logger, database db.Database) *UserHandler {
	return &UserHandler{
		Handler: NewHandler(logger),
		db:      database,
	}
}

type CreateUserRequest struct {
	Email string `json:"email"`
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := validator.ValidateEmail(req.Email); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error(), map[string]string{
			"field": "email",
			"value": req.Email,
		})
		return
	}

	user, err := h.db.CreateUser(r.Context(), req.Email)
	if err != nil {
		if err == db.ErrDuplicateEmail {
			h.respondWithError(w, http.StatusConflict, "Email already exists", nil)
			return
		}
		h.Logger.Error().Err(err).Msg("Failed to create user")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create user", nil)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, user)
}
