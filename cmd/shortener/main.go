package main

import (
	"net/http"

	"github.com/hydra13/shortify/internal/handler"
	db "github.com/hydra13/shortify/internal/repository/inmemory_db"
)

func main() {
	repo := db.New()
	mux := http.NewServeMux()

	getShortURLHandler := handler.CreateGetShortURLHandler(repo)
	getLongURLHandler := handler.CreateGetLongURLHandler(repo)

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
