//go:generate minimock -i .Generator,.UrlsKeeper,.URLValidator -o mocks -s _mock.go -g
package shorter

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/hydra13/shortify/internal/models"
	repository "github.com/hydra13/shortify/internal/repositories"
	authContext "github.com/hydra13/shortify/internal/services/auth_context"
)

type Generator interface {
	GenerateShortID(url string) string
}

type UrlsKeeper interface {
	Save(ctx context.Context, originalURL string, shortURL string, userID string) error
	SaveBatch(ctx context.Context, urls map[string]string, userID string) error
	GetShortURL(ctx context.Context, originalURL string) (string, error)
}

type URLValidator interface {
	Validate(url string) bool
}

type Shorter struct {
	validator URLValidator
	keeper    UrlsKeeper
	generator Generator
	baseURL   string
	log       zerolog.Logger
}

func New(
	validator URLValidator,
	keeper UrlsKeeper,
	generator Generator,
	baseURL string,
	log zerolog.Logger,
) *Shorter {
	return &Shorter{
		validator: validator,
		keeper:    keeper,
		generator: generator,
		baseURL:   baseURL,
		log:       log,
	}
}

func (s Shorter) Create(ctx context.Context, long string) (string, error) {
	if !s.validator.Validate(long) {
		return "", models.ErrValidation
	}

	userID, _ := authContext.GetUserIDFromContext(ctx)
	shortID := s.generator.GenerateShortID(long)
	shortURL := fmt.Sprintf(`%s/%s`, s.baseURL, shortID)

	err := s.keeper.Save(ctx, long, shortID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			shortID, err = s.keeper.GetShortURL(ctx, long)
			if err != nil {
				s.log.Error().
					Str("long_url", long).
					Str("short_url", shortURL).
					Err(err).
					Msg("Error get shortURL from repository during conflict")

				return "", models.ErrInternal
			}

			return fmt.Sprintf(`%s/%s`, s.baseURL, shortID), models.ErrConflict
		}

		s.log.Error().
			Str("long_url", long).
			Str("short_url", shortURL).
			Err(err).
			Msg("Error save url into repository")

		return "", models.ErrInternal
	}

	return shortURL, err
}

func (s Shorter) CreateBatch(ctx context.Context, longURLs map[string]string) (map[string]string, error) {
	shortURLs := make(map[string]string, len(longURLs))
	pairURLs := make(map[string]string, len(longURLs))

	for correlationID, longURL := range longURLs {
		if !s.validator.Validate(longURL) {
			return nil, models.ErrValidation
		}

		shortID := s.generator.GenerateShortID(longURL)
		shortURL := fmt.Sprintf(`%s/%s`, s.baseURL, shortID)

		shortURLs[correlationID] = shortURL
		pairURLs[shortID] = longURL
	}

	userID, _ := authContext.GetUserIDFromContext(ctx)
	err := s.keeper.SaveBatch(ctx, pairURLs, userID)
	if err != nil {
		s.log.Error().
			Interface("pair_urls", pairURLs).
			Err(err).
			Msg("Error save urls into repository")

		return nil, models.ErrInternal
	}

	return shortURLs, nil
}
