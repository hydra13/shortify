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

func (UrlsKeeper *UrlsKeeper) Save(ctx context.Context, originalURL, shortURL string) error {
	return UrlsKeeper.repo.Add(ctx, originalURL, shortURL)
}

func (UrlsKeeper *UrlsKeeper) Get(
	ctx context.Context,
	shortURL string,
) (
	originalURL string,
	found bool,
	err error,
) {
	originalURL, err = UrlsKeeper.repo.Get(ctx, shortURL)

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
