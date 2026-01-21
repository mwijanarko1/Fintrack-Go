package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"fintrack-go/internal/db"
	"fintrack-go/internal/validator"
)

type TransactionHandler struct {
	*Handler
	db db.Database
}

func NewTransactionHandler(logger zerolog.Logger, database db.Database) *TransactionHandler {
	return &TransactionHandler{
		Handler: NewHandler(logger),
		db:      database,
	}
}

type CreateTransactionRequest struct {
	UserID      string     `json:"user_id"`
	CategoryID  *string    `json:"category_id,omitempty"`
	Amount      float64    `json:"amount"`
	Description *string    `json:"description,omitempty"`
	OccurredAt  *time.Time `json:"occurred_at,omitempty"`
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
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

	if req.CategoryID != nil {
		if err := validator.ValidateUUID(*req.CategoryID); err != nil {
			h.respondWithError(w, http.StatusBadRequest, err.Error(), map[string]string{
				"field": "category_id",
				"value": *req.CategoryID,
			})
			return
		}
	}

	if err := validator.ValidateAmount(req.Amount); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error(), map[string]any{
			"field": "amount",
			"value": req.Amount,
		})
		return
	}

	if err := validator.ValidateDescription(req.Description); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error(), map[string]string{
			"field": "description",
			"value": "",
		})
		return
	}

	occurredAt := time.Now()
	if req.OccurredAt != nil {
		occurredAt = *req.OccurredAt
	}

	transaction, err := h.db.CreateTransaction(r.Context(), req.UserID, req.CategoryID, req.Amount, req.Description, occurredAt)
	if err != nil {
		h.Logger.Error().Err(err).Msg("Failed to create transaction")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create transaction", nil)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, transaction)
}

func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
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

	var from, to *time.Time
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Invalid 'from' date format. Use RFC3339", nil)
			return
		}
		from = &t
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Invalid 'to' date format. Use RFC3339", nil)
			return
		}
		to = &t
	}

	if err := validator.ValidateDateRange(from, to); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	transactions, err := h.db.ListTransactions(r.Context(), userID, from, to)
	if err != nil {
		h.Logger.Error().Err(err).Msg("Failed to list transactions")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list transactions", nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, transactions)
}
