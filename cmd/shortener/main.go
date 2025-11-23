package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"

	_ "github.com/hydra13/shortify"
	"github.com/hydra13/shortify/internal/config"
	dbConfig "github.com/hydra13/shortify/internal/config/db"
	longUrlHandler "github.com/hydra13/shortify/internal/handlers/get_long_url_handler"
	shortUrlByJsonHandler "github.com/hydra13/shortify/internal/handlers/get_short_url_by_json_handler"
	shortUrlHandler "github.com/hydra13/shortify/internal/handlers/get_short_url_handler"
	shortUrlsBatchHandler "github.com/hydra13/shortify/internal/handlers/get_short_urls_batch_handler"
	ping "github.com/hydra13/shortify/internal/handlers/ping_handler"
	"github.com/hydra13/shortify/internal/middlewares/compresser"
	"github.com/hydra13/shortify/internal/middlewares/logger"
	gen "github.com/hydra13/shortify/internal/services/short_id_generator"
	shorter "github.com/hydra13/shortify/internal/services/shorter"
	validator "github.com/hydra13/shortify/internal/services/url_validator"
	urlsKeeper "github.com/hydra13/shortify/internal/services/urls_keeper"
)

func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	conf := config.NewConfig()
	conf.ParseConfig()

	dbInstance, err := sql.Open(conf.DatabaseDriver, conf.DatabaseDSN)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to db")
	}
	defer func() {
		if dbInstance != nil {
			dbInstance.Close()
		}
	}()

	repo, err := dbConfig.GetRepository(dbInstance, conf, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init repo")
	}

	generator := gen.New()
	urlValidator := validator.New()
	uk := urlsKeeper.New(repo)
	s := shorter.New(urlValidator, uk, generator, conf.BaseURL, log)

	getLongURLHandler := longUrlHandler.CreateHandler(uk, log)
	getShortURLHandler := shortUrlHandler.CreateHandler(s, conf.BaseURL, log)
	getShortURLbyJSONHandler := shortUrlByJsonHandler.CreateHandler(s, conf.BaseURL, log)
	getShortURLSBatchHandler := shortUrlsBatchHandler.CreateHandler(s, conf.BaseURL, log)

	r := chi.NewRouter()

	r.Use(compresser.CompresserMiddleware)
	r.Use(logger.NewLoggerMiddleware(log))

	r.Post("/", getShortURLHandler)
	if dbInstance != nil {
		pingHandler := ping.CreateHandler(dbInstance, log)
		r.Get("/ping", pingHandler)
	}
	r.Get("/{id}", getLongURLHandler)

	r.Route("/api/shorten", func(r chi.Router) {
		r.Post("/", getShortURLbyJSONHandler)
		r.Post("/batch", getShortURLSBatchHandler)
	})

	log.Fatal().Err(http.ListenAndServe(conf.ServerAddr, r)).Msg("exit")
}
