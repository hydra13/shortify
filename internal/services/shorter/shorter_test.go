package shorter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hydra13/shortify/internal/models"
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
	generator := GeneratorMock{
		wantInput: "https://ya.ru",
		t:         t,
	}

	validator := ValidatorMock{}
	keeper := KeeperMock{}
	baseURL := "http://localhost:8080"

	tests := []struct {
		name      string
		generator Generator
		long      string
		want      string
		err       error
	}{
		{
			name:      "Success",
			generator: generator,
			long:      "https://ya.ru",
			want:      "http://localhost:8080/testing1",
			err:       nil,
		},
		{
			name:      "Error validation",
			generator: generator,
			long:      "",
			want:      "",
			err:       models.ErrValidation,
		},
		{
			name:      "Error validation",
			generator: generator,
			long:      "not-url",
			want:      "",
			err:       models.ErrValidation,
		},
		{
			name: "Error internal during save new short url",
			generator: GeneratorMock{
				wantInput: "https://google.com",
				t:         t,
			},
			long: "https://google.com",
			want: "",
			err:  models.ErrInternal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(validator, keeper, tt.generator, baseURL)
			got, gotErr := s.Create(context.Background(), tt.long)

			if tt.err != nil {
				assert.ErrorIs(t, gotErr, tt.err)

				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
