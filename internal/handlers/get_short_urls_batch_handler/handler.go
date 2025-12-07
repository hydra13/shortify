//go:generate minimock -i .Shorter -o mocks -s _mock.go -g
package getshorturlsbatchhandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/hydra13/shortify/internal/models"
)

type Shorter interface {
	CreateBatch(ctx context.Context, longURLs map[string]string) (map[string]string, error)
}

type RequestRecord struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ResponseRecord struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type Handler struct {
	shorter Shorter
	log     zerolog.Logger
}

func NewHandler(
	shorter Shorter,
	log zerolog.Logger,
) *Handler {
	return &Handler{
		shorter: shorter,
		log:     log,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	var req []RequestRecord
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.log.Debug().
			Err(err).
			Msg("GetShortUrlsBatchHandler: error read request body")

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURLs, err := h.shorter.CreateBatch(r.Context(), toMap(req))
	if err != nil {
		switch err {
		case models.ErrValidation:
			h.log.Debug().
				Interface("input_urls", req).
				Msg("GetShortUrlsBatchHandler: validation error")

			w.WriteHeader(http.StatusBadRequest)
		case models.ErrInternal:
			h.log.Debug().
				Interface("input_urls", req).
				Msg("GetShortUrlsBatchHandler: save url into repository error")

			w.WriteHeader(http.StatusInternalServerError)
		default:
			h.log.Error().
				Interface("input_urls", req).
				Err(err).
				Msg("GetShortUrlsBatchHandler: unhandled error")
		}

		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(toResponse(shortURLs))
	if err != nil {
		log.Error().
			Err(err).
			Msg("GetShortUrlsBatchHandler: error encode response")
	}
}

func toMap(req []RequestRecord) map[string]string {
	m := make(map[string]string, len(req))

	for _, r := range req {
		m[r.CorrelationID] = r.OriginalURL
	}

	return m
}

func toResponse(records map[string]string) []ResponseRecord {
	result := make([]ResponseRecord, 0, len(records))

	for correlationID, shortURL := range records {
		result = append(result, ResponseRecord{
			CorrelationID: correlationID,
			ShortURL:      shortURL,
		})
	}

	return result
}
