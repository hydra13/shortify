package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	longUrlHandler "github.com/hydra13/shortify/internal/handlers/get_long_url_handler"
	shortUrlByJsonHandler "github.com/hydra13/shortify/internal/handlers/get_short_url_by_json_handler"
	shortUrlHandler "github.com/hydra13/shortify/internal/handlers/get_short_url_handler"
	"github.com/hydra13/shortify/internal/logger"
	db "github.com/hydra13/shortify/internal/repositories/inmemory_db"
	gen "github.com/hydra13/shortify/internal/services/short_id_generator"
	shorter "github.com/hydra13/shortify/internal/services/shorter"
	validator "github.com/hydra13/shortify/internal/services/url_validator"
	urlsKeeper "github.com/hydra13/shortify/internal/services/urls_keeper"
)

var (
	serverAddr string
	baseURL    string = "http://localhost:8080"
)

func main() {
	parseFlags()
	parseEnv()

	repo := db.New()
	generator := gen.New()
	urlValidator := validator.New()
	uk := urlsKeeper.New(repo)
	s := shorter.New(urlValidator, uk, generator, baseURL)

	getLongURLHandler := longUrlHandler.CreateHandler(uk)
	getShortURLHandler := shortUrlHandler.CreateHandler(s, baseURL)
	getShortURLbyJSONHandler := shortUrlByJsonHandler.CreateHandler(s, baseURL)

	r := chi.NewRouter()

	r.Use(logger.LoggerMiddleware)

	r.Post("/", getShortURLHandler)
	r.Get("/{id}", getLongURLHandler)

	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", getShortURLbyJSONHandler)
	})

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

func parseEnv() {
	if addr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		serverAddr = addr
	}

	if url, ok := os.LookupEnv("BASE_URL"); ok {
		baseURL = url
	}
}
