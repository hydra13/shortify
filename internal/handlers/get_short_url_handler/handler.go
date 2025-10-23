package getshorturlhandler

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type Generator interface {
	GenerateShortID(url string) string
}

type UrlsKeeper interface {
	Save(ctx context.Context, originalURL string, shortURL string) error
}

func CreateHandler(urlsKeeper UrlsKeeper, generator Generator, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("Error read request body: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		url := string(body)

		if !validation(url) {
			fmt.Printf("Error validation input data: %v\n", url)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		key := generator.GenerateShortID(url)
		err = urlsKeeper.Save(r.Context(), key, url)
		if err != nil {
			fmt.Printf("Error save url into repository: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		shortURL := fmt.Sprintf(`%s/%s`, baseURL, key)

		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}
