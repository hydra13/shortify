//go:generate minimock -i .Shorter,.AuthService -o mocks -s _mock.go -g
package getshorturlbyjsonhandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/hydra13/shortify/internal/models"
	authContext "github.com/hydra13/shortify/internal/services/auth_context"
)

type Shorter interface {
	Create(ctx context.Context, long string) (string, error)
}

type AuthService interface {
	SetAuthCookie(w http.ResponseWriter, userID string)
}

type JSONRequest struct {
	URL string `json:"url"`
}

type JSONResponse struct {
	Result string `json:"result"`
}

func CreateHandler(shorter Shorter, auth AuthService, baseURL string, log zerolog.Logger) http.HandlerFunc {
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

		statusCode := http.StatusCreated
		shortURL, err := shorter.Create(r.Context(), req.URL)
		if err != nil {
			if !errors.Is(err, models.ErrConflict) {
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

			log.Debug().
				Str("input_url", req.URL).
				Msg("GetShortUrlByJsonHandler: url already exist")

			statusCode = http.StatusConflict
		} else {
			userID, isNew := authContext.GetUserIDFromContext(r.Context())
			if isNew {
				auth.SetAuthCookie(w, userID)
			}
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		err = json.NewEncoder(w).Encode(JSONResponse{Result: shortURL})
		if err != nil {
			log.Error().
				Err(err).
				Msg("GetShortUrlByJsonHandler: error encode response")
		}
	}
}
