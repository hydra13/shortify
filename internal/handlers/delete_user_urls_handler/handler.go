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

type Handler struct {
	urlsKeeper UrlsKeeper
	log        zerolog.Logger
}

func NewHandler(urlsKeeper UrlsKeeper, log zerolog.Logger) *Handler {
	return &Handler{
		urlsKeeper: urlsKeeper,
		log:        log,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	var req []string
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.log.Debug().
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

	h.urlsKeeper.DeleteAsync(r.Context(), userID, req)

	w.WriteHeader(http.StatusAccepted)
}
