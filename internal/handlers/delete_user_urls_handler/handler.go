//go:generate minimock -i .UrlsKeeper -o mocks -s _mock.go -g
package deleteuserurlshandler

import (
	"context"
	"encoding/json"
	"net/http"

	authContext "github.com/hydra13/shortify/internal/services/auth_context"
	"github.com/rs/zerolog"
)

type UrlsKeeper interface {
	DeleteAsync(ctx context.Context, userID string, shortURLs []string)
}

func CreateHandler(urlsKeeper UrlsKeeper, log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req []string
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Debug().
				Err(err).
				Msg("DeleteUserUrlsHandler: error read request body")

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userID, isNew := authContext.GetUserIDFromContext(r.Context())

		if isNew {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		urlsKeeper.DeleteAsync(r.Context(), userID, req)

		w.WriteHeader(http.StatusAccepted)
	}
}
