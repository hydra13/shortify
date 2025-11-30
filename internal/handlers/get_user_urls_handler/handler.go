//go:generate minimock -i .UrlsKeeper,.AuthService -o mocks -s _mock.go -g
package getuserurlshandler

import (
	"context"
	"encoding/json"
	"net/http"

	authContext "github.com/hydra13/shortify/internal/services/auth_context"
	"github.com/rs/zerolog"
)

type UrlsKeeper interface {
	GetAllByUser(ctx context.Context, userID string) (urls map[string]string, err error)
}

type ResponseRecord struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func CreateHandler(urlsKeeper UrlsKeeper, log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, isNew := authContext.GetUserIDFromContext(r.Context())
		if isNew {
			log.Debug().
				Msg("GetUserURLsHandler: got request without correct cookie")

			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		urls, err := urlsKeeper.GetAllByUser(r.Context(), userID)
		if err != nil {
			log.Error().
				Str("user_id", userID).
				Err(err).
				Msg("GetUserURLsHandler: error get urls by user")

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		if len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(toResponse(urls))
		if err != nil {
			log.Error().
				Err(err).
				Msg("GetUserURLsHandler: error encode response")
		}
	}
}

func toResponse(urls map[string]string) []ResponseRecord {
	result := make([]ResponseRecord, 0, len(urls))

	for shortURL, originalURL := range urls {
		result = append(result, ResponseRecord{
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		})
	}

	return result
}
