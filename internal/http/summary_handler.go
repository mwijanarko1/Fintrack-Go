package http

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"fintrack-go/internal/db"
	"fintrack-go/internal/validator"
)

type SummaryHandler struct {
	*Handler
	db db.Database
}

func NewSummaryHandler(logger zerolog.Logger, database db.Database) *SummaryHandler {
	return &SummaryHandler{
		Handler: NewHandler(logger),
		db:      database,
	}
}

func (h *SummaryHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
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

	summary, err := h.db.GetSummary(r.Context(), userID, from, to)
	if err != nil {
		h.Logger.Error().Err(err).Msg("Failed to get summary")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get summary", nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, summary)
}
