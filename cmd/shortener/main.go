package main

import (
	"database/sql"
	"flag"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"

	longUrlHandler "github.com/hydra13/shortify/internal/handlers/get_long_url_handler"
	shortUrlByJsonHandler "github.com/hydra13/shortify/internal/handlers/get_short_url_by_json_handler"
	shortUrlHandler "github.com/hydra13/shortify/internal/handlers/get_short_url_handler"
	ping "github.com/hydra13/shortify/internal/handlers/ping_handler"
	"github.com/hydra13/shortify/internal/middlewares/compresser"
	"github.com/hydra13/shortify/internal/middlewares/logger"
	db "github.com/hydra13/shortify/internal/repositories/file_storage"
	gen "github.com/hydra13/shortify/internal/services/short_id_generator"
	shorter "github.com/hydra13/shortify/internal/services/shorter"
	validator "github.com/hydra13/shortify/internal/services/url_validator"
	urlsKeeper "github.com/hydra13/shortify/internal/services/urls_keeper"
)

var (
	serverAddr  string
	baseURL     string = "http://localhost:8080"
	fileStorage string = "./storage.json"
	dbDSN       string = "postgresql://postgres:postgres@localhost:5432/shortify_data?sslmode=disable"
	dbDriver    string = "pgx"
)

func main() {
	parseConfigs()
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	dbInstance, err := sql.Open(dbDriver, dbDSN)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to db")
	}
	defer func() {
		if dbInstance != nil {
			dbInstance.Close()
		}
	}()

	repo, err := db.New(fileStorage, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init repo")
	}

	generator := gen.New()
	urlValidator := validator.New()
	uk := urlsKeeper.New(repo)
	s := shorter.New(urlValidator, uk, generator, baseURL, log)

	getLongURLHandler := longUrlHandler.CreateHandler(uk, log)
	getShortURLHandler := shortUrlHandler.CreateHandler(s, baseURL, log)
	getShortURLbyJSONHandler := shortUrlByJsonHandler.CreateHandler(s, baseURL, log)

	r := chi.NewRouter()

	r.Use(compresser.CompresserMiddleware)
	r.Use(logger.NewLoggerMiddleware(log))

	r.Post("/", getShortURLHandler)
	if dbInstance != nil {
		pingHandler := ping.CreateHandler(dbInstance, log)
		r.Get("/ping", pingHandler)
	}
	r.Get("/{id}", getLongURLHandler)

	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", getShortURLbyJSONHandler)
	})

	http.ListenAndServe(serverAddr, r)
}

func parseConfigs() {
	flag.StringVar(&serverAddr, "a", ":8080", "server address")
	flag.Func("f", "file storage path (default: \"./storage.json\")", func(path string) error {
		if len(path) == 0 {
			return nil
		}

		fileStorage = path

		return nil
	})
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
	flag.Func("d", "database DSN (default: \"postgresql://postgres:postgres@localhost:5432/shortify_data?sslmode=disable\")", func(dsn string) error {
		if len(dsn) == 0 {
			return nil
		}

		dbDSN = dsn

		return nil
	})

	flag.Parse()

	if addr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		serverAddr = addr
	}

	if url, ok := os.LookupEnv("BASE_URL"); ok {
		baseURL = url
	}

	if filePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		fileStorage = filePath
	}

	if dsn, ok := os.LookupEnv("DATABASE_DSN"); ok {
		dbDSN = dsn
	}
}
