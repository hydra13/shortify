package pgstorage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

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

	return dbs, nil
}

func (dbs *DBStorage) Add(ctx context.Context, shortURL string, originalURL string, userID string) error {
	res, err := dbs.db.ExecContext(
		ctx,
		`INSERT INTO shortify_urls (short_url, original_url, user_id)
		VALUES ($1, $2, $3)`,
		shortURL,
		originalURL,
		userID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return repository.ErrConflict
		}

		log.
			Err(err).
			Str("short_url", shortURL).
			Str("original_url", originalURL).
			Msg("DBStorage: error inserting data to db")

		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (dbs *DBStorage) AddBatch(ctx context.Context, batch map[string]string, userID string) error {
	tx, err := dbs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for shortURL, originalURL := range batch {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO shortify_urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
			shortURL,
			originalURL,
			userID,
		)
		if err != nil {
			log.
				Err(err).
				Str("short_url", shortURL).
				Str("original_url", originalURL).
				Msg("DBStorage: error inserting batch data to db")

			return err
		}
	}

	return tx.Commit()
}

func (dbs *DBStorage) Get(ctx context.Context, key string) (string, error) {
	row := dbs.db.QueryRowContext(
		ctx,
		"SELECT original_url, is_deleted FROM shortify_urls WHERE short_url = $1 LIMIT 1",
		key,
	)

	var (
		originalURL string
		isDeleted   bool
	)

	if err := row.Scan(&originalURL, &isDeleted); err != nil {
		log.
			Err(err).
			Msg("DBStorage: error scanning row")
		return "", err
	}

	if isDeleted {
		return "", repository.ErrNotAvailable
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
	defer rows.Close()

	result := make(map[string]string, 0)
	for rows.Next() {
		var shortURL, originalURL string

		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			log.
				Err(err).
				Msg("DBStorage: error scanning row from rows")

			return nil, err
		}

		result[shortURL] = originalURL
	}

	if rows.Err() != nil {
		log.
			Err(err).
			Msg("DBStorage: error iterating over rows")
		return nil, rows.Err()
	}

	return result, nil
}

func (dbs *DBStorage) GetAllByUser(ctx context.Context, userID string) (map[string]string, error) {
	rows, err := dbs.db.QueryContext(
		ctx,
		"SELECT short_url, original_url FROM shortify_urls WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string, 0)
	for rows.Next() {
		var shortURL, originalURL string

		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			log.
				Err(err).
				Msg("DBStorage: error scanning row from rows")

			return nil, err
		}

		result[shortURL] = originalURL
	}

	if rows.Err() != nil {
		log.
			Err(err).
			Msg("DBStorage: error iterating over rows")
		return nil, rows.Err()
	}

	return result, nil
}

func (dbs *DBStorage) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	row := dbs.db.QueryRowContext(
		ctx,
		"SELECT short_url FROM shortify_urls WHERE original_url = $1 LIMIT 1",
		originalURL,
	)

	var shortURL string
	if err := row.Scan(&shortURL); err != nil {
		log.
			Err(err).
			Msg("DBStorage: error scanning row with shortURL")
		return "", err
	}
	return shortURL, nil
}

func (dbs *DBStorage) Delete(ctx context.Context, key string) error {
	_, err := dbs.db.ExecContext(
		ctx,
		"DELETE FROM shortify_urls WHERE short_url = $1",
		key,
	)
	if err != nil {
		log.
			Err(err).
			Msg("DBStorage: error deleting data from db")
	}
	return err
}

func (dbs *DBStorage) DeleteBatch(ctx context.Context, deleteBatch map[string][]string) {
	tx, err := dbs.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()

	for userID, shortURLs := range deleteBatch {
		_, err = tx.ExecContext(
			ctx,
			"UPDATE shortify_urls SET is_deleted = true WHERE user_id = $1 AND short_url = ANY($2)",
			userID,
			shortURLs,
		)
		if err != nil {
			return
		}
	}

	tx.Commit()
}

func (dbs *DBStorage) GetStats(ctx context.Context) (urls int, users int, err error) {
	row := dbs.db.QueryRowContext(
		ctx,
		"SELECT COUNT(DISTINCT short_url) as urls, COUNT(DISTINCT user_id) as users FROM shortify_urls",
	)
	err = row.Scan(&urls, &users)
	return
}
