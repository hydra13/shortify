package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	longUrlHandler "github.com/hydra13/shortify/internal/handlers/get_long_url_handler"
	shortUrlHandler "github.com/hydra13/shortify/internal/handlers/get_short_url_handler"
	db "github.com/hydra13/shortify/internal/repositories/inmemory_db"
	gen "github.com/hydra13/shortify/internal/services/short_id_generator"
)

var serverAddr string
var baseUrl string = "http://localhost:8080"

func main() {
	parseFlags()

	repo := db.New()
	generator := gen.Generator{}

	getShortURLHandler := shortUrlHandler.CreateHandler(repo, generator, baseUrl)
	getLongURLHandler := longUrlHandler.CreateHandler(repo)

	r := chi.NewRouter()

	r.Post("/", getShortURLHandler)
	r.Get("/{id}", getLongURLHandler)

	http.ListenAndServe(serverAddr, r)
}

func parseFlags() {
	flag.StringVar(&serverAddr, "addr", ":8080", "server address")
	flag.Func("base", "result base url (default: \"http://localhost:8080\")", func(url string) error {
		if len(url) == 0 {
			return errors.New("empty base url")
		}

		baseUrl = url

		if baseUrl[len(baseUrl)-1] == '/' {
			baseUrl = baseUrl[:len(baseUrl)-1]
		}

		return nil
	})

	flag.Parse()

	fmt.Println(baseUrl)
}
