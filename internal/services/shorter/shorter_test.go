package shorter

import (
	"bytes"
	"context"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/hydra13/shortify/internal/models"
	"github.com/hydra13/shortify/internal/services/shorter/mocks"
)

type GeneratorMock struct {
	wantInput string
	t         *testing.T
}

func (g GeneratorMock) GenerateShortID(in string) string {
	assert.Equal(g.t, g.wantInput, in)
	return "testing1"
}

type ValidatorMock struct{}

func (v ValidatorMock) Validate(s string) bool {
	return s != "" && s != "not-url"
}

type KeeperMock struct{}

func (k KeeperMock) Save(_ context.Context, originalURL string, shortURL string) error {
	if originalURL == "https://google.com" {
		return models.ErrInternal
	}

	return nil
}

func TestShorter_Create(t *testing.T) {
	var buf bytes.Buffer
	log := zerolog.New(&buf)
	zerolog.SetGlobalLevel(zerolog.Disabled)

	baseURL := "http://localhost:8080"

	tests := []struct {
		name      string
		generator func(mc *minimock.Controller) Generator
		validator func(mc *minimock.Controller) URLValidator
		keeper    func(mc *minimock.Controller) UrlsKeeper

		long string
		want string
		err  error
	}{
		{
			name: "Success",
			generator: func(mc *minimock.Controller) Generator {
				return mocks.NewGeneratorMock(mc).
					GenerateShortIDMock.
					Expect("https://ya.ru").
					Return("testing1")
			},
			validator: func(mc *minimock.Controller) URLValidator {
				return mocks.NewURLValidatorMock(mc).
					ValidateMock.
					Expect("https://ya.ru").
					Return(true)
			},
			keeper: func(mc *minimock.Controller) UrlsKeeper {
				return mocks.NewUrlsKeeperMock(mc).
					SaveMock.
					Expect(minimock.AnyContext, "https://ya.ru", "testing1").
					Return(nil)
			},
			long: "https://ya.ru",
			want: "http://localhost:8080/testing1",
			err:  nil,
		},
		{
			name:      "Error validation",
			generator: func(mc *minimock.Controller) Generator { return mocks.NewGeneratorMock(mc) },
			validator: func(mc *minimock.Controller) URLValidator {
				return mocks.NewURLValidatorMock(mc).
					ValidateMock.
					Expect("").
					Return(false)
			},
			keeper: func(mc *minimock.Controller) UrlsKeeper { return mocks.NewUrlsKeeperMock(mc) },
			long:   "",
			want:   "",
			err:    models.ErrValidation,
		},
		{
			name:      "Error validation",
			generator: func(mc *minimock.Controller) Generator { return mocks.NewGeneratorMock(mc) },
			validator: func(mc *minimock.Controller) URLValidator {
				return mocks.NewURLValidatorMock(mc).
					ValidateMock.
					Expect("not-url").
					Return(false)
			},
			keeper: func(mc *minimock.Controller) UrlsKeeper { return mocks.NewUrlsKeeperMock(mc) },
			long:   "not-url",
			want:   "",
			err:    models.ErrValidation,
		},
		{
			name: "Error internal during save new short url",
			generator: func(mc *minimock.Controller) Generator {
				return mocks.NewGeneratorMock(mc).
					GenerateShortIDMock.
					Expect("https://google.com").
					Return("testing1")
			},
			validator: func(mc *minimock.Controller) URLValidator {
				return mocks.NewURLValidatorMock(mc).
					ValidateMock.
					Expect("https://google.com").
					Return(true)
			},
			keeper: func(mc *minimock.Controller) UrlsKeeper {
				return mocks.NewUrlsKeeperMock(mc).
					SaveMock.
					Expect(minimock.AnyContext, "https://google.com", "testing1").
					Return(models.ErrInternal)
			},
			long: "https://google.com",
			want: "",
			err:  models.ErrInternal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)

			validator := tt.validator(mc)
			keeper := tt.keeper(mc)
			generator := tt.generator(mc)

			s := New(validator, keeper, generator, baseURL, log)
			got, gotErr := s.Create(context.Background(), tt.long)

			if tt.err != nil {
				assert.ErrorIs(t, gotErr, tt.err)

				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
