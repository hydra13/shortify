//go:generate minimock -i .Repository -o mocks -s _mock.go -g
package statshandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
)

type Repository interface {
	GetStats(ctx context.Context) (urls int, users int, err error)
}

type Handler struct {
	repo Repository
	log  zerolog.Logger
}

type Response struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

func NewHandler(repo Repository, log zerolog.Logger) *Handler {
	return &Handler{
		repo: repo,
		log:  log,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	urls, users, err := h.repo.GetStats(r.Context())
	if err != nil {
		h.log.Error().Err(err).Msg("StatsHandler: error get stats")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{URLs: urls, Users: users})
}
