package urlskeeper

import (
	"context"
	"time"

	"github.com/hydra13/shortify/internal/models"
	repository "github.com/hydra13/shortify/internal/repositories"
)

type URLForDelete struct {
	shortURL string
	userID   string
}

type UrlsKeeper struct {
	repo     repository.Repository
	deleteCh chan URLForDelete
}

func New(ctx context.Context, repo repository.Repository) *UrlsKeeper {
	deleteCh := make(chan URLForDelete, 1024)

	keeper := &UrlsKeeper{
		repo:     repo,
		deleteCh: deleteCh,
	}

	go keeper.Run(ctx)

	return keeper
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

		if err == repository.ErrNotAvailable {
			return "", false, models.ErrURLIsDeleted
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

func (uk *UrlsKeeper) DeleteAsync(ctx context.Context, userID string, shortURLs []string) {
	go func() {
		for _, shortURL := range shortURLs {
			uk.deleteCh <- URLForDelete{
				shortURL: shortURL,
				userID:   userID,
			}
		}
	}()
}

func (uk *UrlsKeeper) Run(ctx context.Context) {
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()

	deleteBatch := make(map[string][]string)

	for {
		select {
		case <-ctx.Done():
			return
		case url := <-uk.deleteCh:
			deleteBatch[url.userID] = append(deleteBatch[url.userID], url.shortURL)
		case <-t.C:
			if len(deleteBatch) > 0 {
				uk.repo.DeleteBatch(ctx, deleteBatch)
				deleteBatch = make(map[string][]string)
			}
		}
	}
}
