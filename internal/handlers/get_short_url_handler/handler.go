package getshorturlhandler

import (
	"context"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/hydra13/shortify/internal/models"
)

type Shorter interface {
	Create(ctx context.Context, long string) (string, error)
}

func CreateHandler(shorter Shorter, baseURL string) http.HandlerFunc {
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

		shortURL, err := shorter.Create(r.Context(), longURL)
		if err != nil {
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

		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(shortURL))
		if err != nil {
			log.Error().
				Err(err).
				Msg("GetShortUrlHandler: error write response")
		}
	}
}
