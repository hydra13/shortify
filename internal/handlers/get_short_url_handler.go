package handler

import (
	"fmt"
	"io"
	"net/http"

	repo "github.com/hydra13/shortify/internal/repositories"
	generator "github.com/hydra13/shortify/internal/services/short_id_generator"
)

func CreateGetShortURLHandler(repository repo.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("Error read request body: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Пока осознанно не делаю проверки на то что получен URL
		url := string(body)
		key := generator.GenerateShortID(url)
		repository.Add(key, url)
		shortURL := fmt.Sprintf(`http://localhost:8080/%s`, key)

		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}
