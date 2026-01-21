package http

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
)

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details any    `json:"details,omitempty"`
	} `json:"error"`
}

type Handler struct {
	Logger zerolog.Logger
}

func NewHandler(logger zerolog.Logger) *Handler {
	return &Handler{
		Logger: logger,
	}
}

func (h *Handler) respondWithError(w http.ResponseWriter, code int, message string, details ...any) {
	errResp := ErrorResponse{}
	errResp.Error.Code = http.StatusText(code)
	errResp.Error.Message = message
	
	if len(details) > 0 {
		errResp.Error.Details = details[0]
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		h.Logger.Error().Err(err).Msg("Failed to encode error response")
	}
}

func (h *Handler) respondWithJSON(w http.ResponseWriter, code int, payload any) {
	if payload == nil {
		w.WriteHeader(code)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.Logger.Error().Err(err).Msg("Failed to encode JSON response")
	}
}
