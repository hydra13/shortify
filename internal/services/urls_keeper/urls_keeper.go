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

func (UrlsKeeper *UrlsKeeper) Save(ctx context.Context, originalURL, shortID string) error {
	return UrlsKeeper.repo.Add(ctx, shortID, originalURL)
}

func (UrlsKeeper *UrlsKeeper) Get(
	ctx context.Context,
	shortID string,
) (
	originalURL string,
	found bool,
	err error,
) {
	originalURL, err = UrlsKeeper.repo.Get(ctx, shortID)
	if err != nil {
		if err == repository.ErrKeyNotFound {
			return "", false, nil
		}

		return "", false, err
	}

	return originalURL, true, nil
}

func (UrlsKeeper *UrlsKeeper) Delete(ctx context.Context, shortURL string) error {
	return UrlsKeeper.repo.Delete(ctx, shortURL)
}
