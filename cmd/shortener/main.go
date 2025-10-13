package main

import (
	"net/http"

	getLongUrlHandler "github.com/hydra13/shortify/internal/handlers/get_long_url_handler"
	getShortUrlHandler "github.com/hydra13/shortify/internal/handlers/get_short_url_handler"
	db "github.com/hydra13/shortify/internal/repositories/inmemory_db"
)

func main() {
	repo := db.New()
	mux := http.NewServeMux()

	getShortURLHandler := getShortUrlHandler.CreateHandler(repo)
	getLongURLHandler := getLongUrlHandler.CreateHandler(repo)

	mainHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			getShortURLHandler(w, r)
		case http.MethodGet:
			getLongURLHandler(w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	mux.HandleFunc(`/`, mainHandler)

	http.ListenAndServe(":8080", mux)
}
