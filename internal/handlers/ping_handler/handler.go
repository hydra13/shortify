//go:generate minimock -i .DB -o mocks -s _mock.go -g
package pinghandler

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type DB interface {
	PingContext(ctx context.Context) error
}

type Handler struct {
	db  DB
	log zerolog.Logger
}

func NewHandler(
	db DB,
	log zerolog.Logger,
) *Handler {
	return &Handler{
		db:  db,
		log: log,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctxTimeout, ctxCancel := context.WithTimeout(context.Background(), time.Second*3)
	defer ctxCancel()

	if err := h.db.PingContext(ctxTimeout); err != nil {
		h.log.Error().Err(err).Msg("Database ping error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
