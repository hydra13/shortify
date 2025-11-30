package urlskeeper

import (
	"context"

	repository "github.com/hydra13/shortify/internal/repositories"
)

type UrlsKeeper struct {
	repo repository.Repository
}

func New(repo repository.Repository) *UrlsKeeper {
	return &UrlsKeeper{
		repo: repo,
	}
}

func (uk *UrlsKeeper) Save(ctx context.Context, originalURL, shortID, userID string) error {
	return uk.repo.Add(ctx, shortID, originalURL, userID)
}

func (uk *UrlsKeeper) SaveBatch(
	ctx context.Context,
	urls map[string]string,
	userID string,
) error {
	return uk.repo.AddBatch(ctx, urls, userID)
}

func (uk *UrlsKeeper) Get(
	ctx context.Context,
	shortID string,
) (
	originalURL string,
	found bool,
	err error,
) {
	originalURL, err = uk.repo.Get(ctx, shortID)
	if err != nil {
		if err == repository.ErrKeyNotFound {
			return "", false, nil
		}

		return "", false, err
	}

	return originalURL, true, nil
}

func (uk *UrlsKeeper) GetAllByUser(ctx context.Context, userID string) (urls map[string]string, err error) {
	return uk.repo.GetAllByUser(ctx, userID)
}

func (uk *UrlsKeeper) Delete(ctx context.Context, shortURL string) error {
	return uk.repo.Delete(ctx, shortURL)
}

func (uk *UrlsKeeper) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	return uk.repo.GetShortURL(ctx, originalURL)
}
