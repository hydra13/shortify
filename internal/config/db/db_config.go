package db

import (
	"database/sql"

	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"

	"github.com/hydra13/shortify/internal/config"
	repository "github.com/hydra13/shortify/internal/repositories"
	dbStorage "github.com/hydra13/shortify/internal/repositories/db_storage"
	fileStorage "github.com/hydra13/shortify/internal/repositories/file_storage"
	inmemorydb "github.com/hydra13/shortify/internal/repositories/inmemory_db"
)

func GetRepository(dbInstance *sql.DB, conf *config.Config, log zerolog.Logger) (repository.Repository, error) {
	switch conf.CurrentMode {
	case config.ModeDBStorage:
		err := applyMigrations(dbInstance, conf.DatabaseDriver)
		if err != nil {
			return nil, err
		}

		repo, err := dbStorage.New(dbInstance, log)
		if err != nil {
			return nil, err
		}

		return repo, nil
	case config.ModeFileStorage:
		repo, err := fileStorage.New(conf.FileStoragePath, log)
		if err != nil {
			return nil, err
		}

		return repo, nil
	default:
	}

	return inmemorydb.New(), nil
}

func applyMigrations(db *sql.DB, dbDriver string) error {
	if err := goose.SetDialect(dbDriver); err != nil {
		return err
	}

	return goose.Up(db, "migrations")
}
