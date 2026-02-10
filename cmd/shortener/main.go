package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"

	_ "github.com/hydra13/shortify"
	"github.com/hydra13/shortify/internal/config"
	dbConfig "github.com/hydra13/shortify/internal/config/db"
	deleteUrlsHandler "github.com/hydra13/shortify/internal/handlers/delete_user_urls_handler"
	longUrlHandler "github.com/hydra13/shortify/internal/handlers/get_long_url_handler"
	shortUrlByJsonHandler "github.com/hydra13/shortify/internal/handlers/get_short_url_by_json_handler"
	shortUrlHandler "github.com/hydra13/shortify/internal/handlers/get_short_url_handler"
	shortUrlsBatchHandler "github.com/hydra13/shortify/internal/handlers/get_short_urls_batch_handler"
	userUrlsHandler "github.com/hydra13/shortify/internal/handlers/get_user_urls_handler"
	ping "github.com/hydra13/shortify/internal/handlers/ping_handler"
	authMiddleware "github.com/hydra13/shortify/internal/middlewares/auth"
	"github.com/hydra13/shortify/internal/middlewares/compresser"
	"github.com/hydra13/shortify/internal/middlewares/logger"
	auditService "github.com/hydra13/shortify/internal/services/audit"
	auditSaverService "github.com/hydra13/shortify/internal/services/audit_saver"
	auditSenderService "github.com/hydra13/shortify/internal/services/audit_sender"
	authService "github.com/hydra13/shortify/internal/services/auth"
	gen "github.com/hydra13/shortify/internal/services/short_id_generator"
	shorter "github.com/hydra13/shortify/internal/services/shorter"
	validator "github.com/hydra13/shortify/internal/services/url_validator"
	urlsKeeper "github.com/hydra13/shortify/internal/services/urls_keeper"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func main() {
	printBuildInfo()
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	uk := urlsKeeper.New(ctx, repo)
	s := shorter.New(urlValidator, uk, generator, conf.BaseURL, log)
	auth := authService.New()
	audit := auditService.New()

	if conf.AuditFile != "" {
		auditSaver := auditSaverService.New(conf.AuditFile, log)
		audit.Subscribe(auditSaver)
	}

	if conf.AuditURL != "" {
		auditSender := auditSenderService.New(conf.AuditURL, log)
		audit.Subscribe(auditSender)
	}

	getLongURLHandler := longUrlHandler.NewHandler(uk, audit, log)
	getShortURLHandler := shortUrlHandler.NewHandler(s, auth, audit, log)
	getShortURLbyJSONHandler := shortUrlByJsonHandler.NewHandler(s, auth, audit, log)
	getShortURLSBatchHandler := shortUrlsBatchHandler.NewHandler(s, log)
	getUserURLSHandler := userUrlsHandler.NewHandler(uk, s, log)
	deleteUserUrlsHandler := deleteUrlsHandler.NewHandler(uk, log)

	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(compresser.CompresserMiddleware)
		r.Use(logger.NewLoggerMiddleware(log))
		r.Use(authMiddleware.NewAuthMiddleware(auth, log))

		r.Post("/", getShortURLHandler.Handle)
		if dbInstance != nil {
			pingHandler := ping.NewHandler(dbInstance, log)
			r.Get("/ping", pingHandler.Handle)
		}
		r.Get("/{id}", getLongURLHandler.Handle)

		r.Route("/api/", func(r chi.Router) {
			r.Route("/shorten", func(r chi.Router) {
				r.Post("/", getShortURLbyJSONHandler.Handle)
				r.Post("/batch", getShortURLSBatchHandler.Handle)
			})
			r.Route("/user/urls", func(r chi.Router) {
				r.Get("/", getUserURLSHandler.Handle)
				r.Delete("/", deleteUserUrlsHandler.Handle)
			})
		})
	})

	if conf.ProfilerEnabled {
		go startPprof(log)
	}

	log.Debug().Msg("starting server at " + conf.ServerAddr)

	log.Fatal().Err(http.ListenAndServe(conf.ServerAddr, r)).Msg("exit")
}

func startPprof(log zerolog.Logger) {
	// доступ только с localhost
	localAddr := "127.0.0.1:6060"

	r := chi.NewRouter()
	r.Mount("/debug", middleware.Profiler())

	log.Debug().Msg("starting pprof at " + localAddr)
	log.Fatal().Err(http.ListenAndServe(localAddr, r)).Msg("pprof exit")
}
