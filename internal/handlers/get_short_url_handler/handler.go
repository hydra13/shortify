//go:generate minimock -i .Shorter -o mocks -s _mock.go -g
package getshorturlhandler

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/hydra13/shortify/internal/models"
)

type Shorter interface {
	Create(ctx context.Context, long string) (string, error)
}

func CreateHandler(shorter Shorter, baseURL string, log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Debug().
				Err(err).
				Msg("GetShortUrlHandler: error read request body")

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		longURL := string(body)

		statusCode := http.StatusCreated
		shortURL, err := shorter.Create(r.Context(), longURL)
		if err != nil {
			if !errors.Is(err, models.ErrConflict) {
				switch err {
				case models.ErrValidation:
					log.Debug().
						Str("input_url", longURL).
						Msg("GetShortUrlHandler: validation error")

					w.WriteHeader(http.StatusBadRequest)
				case models.ErrInternal:
					log.Debug().
						Str("input_url", longURL).
						Msg("GetShortUrlHandler: save url into repository error")

					w.WriteHeader(http.StatusInternalServerError)
				default:
					log.Error().
						Str("input_url", longURL).
						Err(err).
						Msg("GetShortUrlHandler: unhandled error")
				}

				return
			}

			log.Debug().
				Str("input_url", longURL).
				Msg("GetShortUrlHandler: url already exists")

			statusCode = http.StatusConflict
		}

		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(statusCode)
		_, err = w.Write([]byte(shortURL))
		if err != nil {
			log.Error().
				Err(err).
				Msg("GetShortUrlHandler: error write response")
		}
	}
}
