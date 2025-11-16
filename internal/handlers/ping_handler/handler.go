package pinghandler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

func CreateHandler(db *sql.DB, log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctxTimeout, ctxCancel := context.WithTimeout(context.Background(), time.Second*3)
		defer ctxCancel()

		if err := db.PingContext(ctxTimeout); err != nil {
			log.Error().Err(err).Msg("Database ping error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
