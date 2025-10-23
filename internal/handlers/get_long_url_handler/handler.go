package getlongurlhandler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hydra13/shortify/internal/config"
)

type UrlsKeeper interface {
	Get(ctx context.Context, shortURL string) (url string, found bool, err error)
}

func CreateHandler(urlsKeeper UrlsKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Длина пути должна быть равна длине ключа + `/`
		if len(r.URL.Path) != config.KeyLength+1 {
			fmt.Printf("Incorrect request path: %v\n", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		key := r.URL.Path[1:]
		url, found, err := urlsKeeper.Get(r.Context(), key)

		if err != nil {
			fmt.Printf("Error get url from repository: %v\n", err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !found {
			fmt.Printf("Url not found: %v\n", key)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
