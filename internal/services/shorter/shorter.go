//go:generate minimock -i .Generator,.UrlsKeeper,.URLValidator -o mocks -s _mock.go -g
package shorter

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/hydra13/shortify/internal/models"
)

type Generator interface {
	GenerateShortID(url string) string
}

type UrlsKeeper interface {
	Save(ctx context.Context, originalURL string, shortURL string) error
	SaveBatch(ctx context.Context, urls map[string]string) error
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

	shortID := s.generator.GenerateShortID(long)
	shortURL := fmt.Sprintf(`%s/%s`, s.baseURL, shortID)

	err := s.keeper.Save(ctx, long, shortID)
	if err != nil {
		s.log.Error().
			Str("long_url", long).
			Str("short_url", shortURL).
			Err(err).
			Msg("Error save url into repository")

		return "", models.ErrInternal
	}

	return shortURL, nil
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

	err := s.keeper.SaveBatch(ctx, pairURLs)
	if err != nil {
		s.log.Error().
			Interface("pair_urls", pairURLs).
			Err(err).
			Msg("Error save urls into repository")

		return nil, models.ErrInternal
	}

	return shortURLs, nil
}
