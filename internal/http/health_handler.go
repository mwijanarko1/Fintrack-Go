package http

import (
	"net/http"

	"github.com/rs/zerolog"
	"fintrack-go/internal/db"
)

type HealthHandler struct {
	*Handler
	db db.Database
}

func NewHealthHandler(logger zerolog.Logger, database db.Database) *HealthHandler {
	return &HealthHandler{
		Handler: NewHandler(logger),
		db:      database,
	}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if err := h.db.Ping(ctx); err != nil {
		h.Logger.Error().Err(err).Msg("Health check failed")
		h.respondWithError(w, http.StatusServiceUnavailable, "Service unavailable", nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}
