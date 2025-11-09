//go:generate minimock -i .Shorter -o mocks -s _mock.go -g
package getshorturlbyjsonhandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/hydra13/shortify/internal/models"
)

type Shorter interface {
	Create(ctx context.Context, long string) (string, error)
}

type JSONRequest struct {
	URL string `json:"url"`
}

type JSONResponse struct {
	Result string `json:"result"`
}

func CreateHandler(shorter Shorter, baseURL string, log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req JSONRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Debug().
				Err(err).
				Msg("GetShortUrlByJsonHandler: error read request body")

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		shortURL, err := shorter.Create(r.Context(), req.URL)
		if err != nil {
			switch err {
			case models.ErrValidation:
				log.Debug().
					Str("input_url", req.URL).
					Msg("GetShortUrlByJsonHandler: validation error")

				w.WriteHeader(http.StatusBadRequest)
			case models.ErrInternal:
				log.Debug().
					Str("input_url", req.URL).
					Msg("GetShortUrlByJsonHandler: save url into repository error")

				w.WriteHeader(http.StatusInternalServerError)
			default:
				log.Error().
					Str("input_url", req.URL).
					Err(err).
					Msg("GetShortUrlByJsonHandler: unhandled error")
			}

			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(JSONResponse{Result: shortURL})
		if err != nil {
			log.Error().
				Err(err).
				Msg("GetShortUrlByJsonHandler: error encode response")
		}
	}
}
