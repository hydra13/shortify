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
