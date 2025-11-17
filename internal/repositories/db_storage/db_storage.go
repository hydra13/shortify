package pgstorage

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog"

	repository "github.com/hydra13/shortify/internal/repositories"
)

type DB interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type DBStorage struct {
	db  DB
	log zerolog.Logger
}

func New(db DB, log zerolog.Logger) (repository.Repository, error) {
	dbs := &DBStorage{
		db:  db,
		log: log,
	}

	if err := dbs.initDB(); err != nil {
		return nil, err
	}

	return dbs, nil
}

func (dbs *DBStorage) initDB() error {
	res, err := dbs.db.ExecContext(
		context.Background(),
		`CREATE TABLE IF NOT EXISTS shortify_urls (
			id BIGSERIAL PRIMARY KEY,
    		short_url TEXT NOT NULL UNIQUE,
    		original_url TEXT NOT NULL,
    		created_at TIMESTAMP NOT NULL DEFAULT now()
	)`)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (dbs *DBStorage) Add(ctx context.Context, key string, value string) error {
	res, err := dbs.db.ExecContext(
		ctx,
		"INSERT INTO shortify_urls (short_url, original_url) VALUES ($1, $2)",
		key,
		value,
	)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (dbs *DBStorage) AddBatch(ctx context.Context, batch map[string]string) error {
	tx, err := dbs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for shortURL, originalURL := range batch {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO shortify_urls (short_url, original_url) VALUES ($1, $2)",
			shortURL,
			originalURL,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (dbs *DBStorage) Get(ctx context.Context, key string) (string, error) {
	row := dbs.db.QueryRowContext(
		ctx,
		"SELECT original_url FROM shortify_urls WHERE short_url = $1 LIMIT 1",
		key,
	)
	if row.Err() != nil {
		return "", row.Err()
	}

	var originalURL string
	if err := row.Scan(&originalURL); err != nil {
		return "", err
	}

	return originalURL, nil
}

func (dbs *DBStorage) GetAll(ctx context.Context) (map[string]string, error) {
	rows, err := dbs.db.QueryContext(
		ctx,
		"SELECT short_url, original_url FROM shortify_urls",
	)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, 0)
	for rows.Next() {
		var shortURL, originalURL string

		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			return nil, err
		}

		result[shortURL] = originalURL
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

func (dbs *DBStorage) Delete(ctx context.Context, key string) error {
	res, err := dbs.db.ExecContext(
		ctx,
		"DELETE FROM shortify_urls WHERE short_url = $1",
		key,
	)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}
