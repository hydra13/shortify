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
	shortUrlsBatchHandler "github.com/hydra13/shortify/internal/handlers/get_short_urls_batch_handler"
	ping "github.com/hydra13/shortify/internal/handlers/ping_handler"
	"github.com/hydra13/shortify/internal/middlewares/compresser"
	"github.com/hydra13/shortify/internal/middlewares/logger"
	repository "github.com/hydra13/shortify/internal/repositories"
	dbStorage "github.com/hydra13/shortify/internal/repositories/db_storage"
	fileStorage "github.com/hydra13/shortify/internal/repositories/file_storage"
	inmemorydb "github.com/hydra13/shortify/internal/repositories/inmemory_db"
	gen "github.com/hydra13/shortify/internal/services/short_id_generator"
	shorter "github.com/hydra13/shortify/internal/services/shorter"
	validator "github.com/hydra13/shortify/internal/services/url_validator"
	urlsKeeper "github.com/hydra13/shortify/internal/services/urls_keeper"
)

type Mode int

const (
	ModeInMemory Mode = iota + 1
	ModeFileStorage
	ModeDBStorage
)

var (
	serverAddr      string
	baseURL         string = "http://localhost:8080"
	fileStoragePath string = "./storage.json"
	dbDSN           string = "postgresql://postgres:postgres@localhost:5432/shortify_data?sslmode=disable"
	dbDriver        string = "pgx"
	currentMode     Mode   = ModeInMemory
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

	repo, err := getRepository(dbInstance, log)
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
	getShortURLSBatchHandler := shortUrlsBatchHandler.CreateHandler(s, baseURL, log)

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

	http.ListenAndServe(serverAddr, r)
}

func parseConfigs() {
	flag.StringVar(&serverAddr, "a", ":8080", "server address")
	flag.Func("f", "file storage path (default: \"./storage.json\")", func(path string) error {
		if len(path) == 0 {
			return nil
		}

		fileStoragePath = path

		if currentMode <= ModeFileStorage {
			currentMode = ModeFileStorage
		}

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

		if currentMode <= ModeDBStorage {
			currentMode = ModeDBStorage
		}

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
		fileStoragePath = filePath

		if currentMode <= ModeFileStorage {
			currentMode = ModeFileStorage
		}
	}

	if dsn, ok := os.LookupEnv("DATABASE_DSN"); ok {
		dbDSN = dsn

		if currentMode <= ModeDBStorage {
			currentMode = ModeDBStorage
		}
	}
}

func getRepository(dbInstance *sql.DB, log zerolog.Logger) (repository.Repository, error) {
	switch currentMode {
	case ModeDBStorage:
		repo, err := dbStorage.New(dbInstance, log)
		if err != nil {
			return nil, err
		}

		return repo, nil
	case ModeFileStorage:
		repo, err := fileStorage.New(fileStoragePath, log)
		if err != nil {
			return nil, err
		}

		return repo, nil
	default:
	}

	return inmemorydb.New(), nil
}
