package getshorturlhandler

import (
	"fmt"
	"io"
	"net/http"

	repo "github.com/hydra13/shortify/internal/repositories"
)

type Generator interface {
	GenerateShortID(url string) string
}

func CreateHandler(repository repo.Repository, generator Generator) http.HandlerFunc {
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
		repository.Add(key, url)
		shortURL := fmt.Sprintf(`http://localhost:8080/%s`, key)

		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}
