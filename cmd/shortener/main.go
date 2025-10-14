package main

import (
	"flag"
	"net/http"

	"github.com/go-chi/chi/v5"

	longUrlHandler "github.com/hydra13/shortify/internal/handlers/get_long_url_handler"
	shortUrlHandler "github.com/hydra13/shortify/internal/handlers/get_short_url_handler"
	db "github.com/hydra13/shortify/internal/repositories/inmemory_db"
	gen "github.com/hydra13/shortify/internal/services/short_id_generator"
)

var serverAddr string
var baseURL string = "http://localhost:8080"

func main() {
	parseFlags()

	repo := db.New()
	generator := gen.Generator{}

	getShortURLHandler := shortUrlHandler.CreateHandler(repo, generator, baseURL)
	getLongURLHandler := longUrlHandler.CreateHandler(repo)

	r := chi.NewRouter()

	r.Post("/", getShortURLHandler)
	r.Get("/{id}", getLongURLHandler)

	http.ListenAndServe(serverAddr, r)
}

func parseFlags() {
	flag.StringVar(&serverAddr, "a", ":8080", "server address")
	flag.Func("b", "result base url (default: \"http://localhost:8080\")", func(url string) error {
		if len(url) == 0 {
			return nil
		}

		baseURL = url

		if baseURL[len(baseURL)-1] == '/' {
			baseURL = baseURL[:len(baseURL)-1]
		}

		return nil
	})

	flag.Parse()
}
