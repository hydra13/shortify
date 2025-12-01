//go:generate minimock -i .UrlsKeeper -o mocks -s _mock.go -g
package getlongurlhandler

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/hydra13/shortify/internal/config"
	"github.com/hydra13/shortify/internal/models"
)

type UrlsKeeper interface {
	Get(ctx context.Context, shortURL string) (url string, found bool, err error)
}

func CreateHandler(urlsKeeper UrlsKeeper, log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Длина пути должна быть равна длине ключа + `/`
		if len(r.URL.Path) != config.KeyLength+1 {
			log.Debug().
				Str("path", r.URL.Path).
				Msg("GetLongUrlHandler: incorrect request path")

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		key := r.URL.Path[1:]
		url, found, err := urlsKeeper.Get(r.Context(), key)
		if err != nil {
			switch err {
			case models.ErrURLIsDeleted:
				log.Debug().
					Str("short_url_key", key).
					Msg("GetLongUrlHandler: error get deleted url")

				w.WriteHeader(http.StatusGone)
			default:
				log.Error().
					Err(err).
					Str("short_url_key", key).
					Msg("GetLongUrlHandler: error get url from repository")

				w.WriteHeader(http.StatusInternalServerError)
			}

			return
		}

		if !found {
			log.Debug().
				Str("short_url_key", key).
				Msg("GetLongUrlHandler: url not found")

			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
