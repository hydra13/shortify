package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	longUrlHandler "github.com/hydra13/shortify/internal/handlers/get_long_url_handler"
	shortUrlHandler "github.com/hydra13/shortify/internal/handlers/get_short_url_handler"
	db "github.com/hydra13/shortify/internal/repositories/inmemory_db"
	gen "github.com/hydra13/shortify/internal/services/short_id_generator"
)

func main() {
	repo := db.New()
	generator := gen.Generator{}

	getShortURLHandler := shortUrlHandler.CreateHandler(repo, generator)
	getLongURLHandler := longUrlHandler.CreateHandler(repo)

	r := chi.NewRouter()

	r.Post("/", getShortURLHandler)
	r.Get("/{id}", getLongURLHandler)

	http.ListenAndServe(":8080", r)
}
