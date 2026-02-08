//go:generate minimock -i .Shorter,.AuthService,.AuditService -o mocks -s _mock.go -g
package getshorturlhandler

import (
	"context"
	"errors"
	"io"
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

type AuditService interface {
	PublishShortenEvent(longURL, userID string)
}

type Handler struct {
	shorter Shorter
	auth    AuthService
	audit   AuditService
	log     zerolog.Logger
}

func NewHandler(
	shorter Shorter,
	auth AuthService,
	audit AuditService,
	log zerolog.Logger,
) *Handler {
	return &Handler{
		shorter: shorter,
		auth:    auth,
		audit:   audit,
		log:     log,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Debug().
			Err(err).
			Msg("GetShortUrlHandler: error read request body")

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	longURL := string(body)

	statusCode := http.StatusCreated
	shortURL, err := h.shorter.Create(r.Context(), longURL)
	if err != nil {
		if !errors.Is(err, models.ErrConflict) {
			switch err {
			case models.ErrValidation:
				h.log.Debug().
					Str("input_url", longURL).
					Msg("GetShortUrlHandler: validation error")

				w.WriteHeader(http.StatusBadRequest)
			case models.ErrInternal:
				h.log.Debug().
					Str("input_url", longURL).
					Msg("GetShortUrlHandler: save url into repository error")

				w.WriteHeader(http.StatusInternalServerError)
			default:
				h.log.Error().
					Str("input_url", longURL).
					Err(err).
					Msg("GetShortUrlHandler: unhandled error")
			}

			return
		}

		h.log.Debug().
			Str("input_url", longURL).
			Msg("GetShortUrlHandler: url already exists")

		statusCode = http.StatusConflict
	} else {
		userID, isNew := authContext.GetUserIDFromContext(r.Context())
		if isNew {
			h.auth.SetAuthCookie(w, userID)
		}

		h.audit.PublishShortenEvent(longURL, userID)
	}

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		h.log.Error().
			Err(err).
			Msg("GetShortUrlHandler: error write response")
	}
}
