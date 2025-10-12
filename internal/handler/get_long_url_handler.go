package handler

import (
	"fmt"
	"net/http"

	"github.com/hydra13/shortify/internal/config"
	repo "github.com/hydra13/shortify/internal/repository"
)

func CreateGetLongURLHandler(repository repo.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Длина пути должна быть равна длине ключа + `/`
		if len(r.URL.Path) != config.KeyLength+1 {
			fmt.Printf("Incorrect request path: %v\n", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		key := r.URL.Path[1:]
		url, err := repository.Get(key)

		if err != nil {
			if err == repo.ErrKeyNotFound {
				fmt.Printf("Url not found: %v\n", key)
			} else {
				fmt.Printf("Error get url from repository: %v\n", err)
			}

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
