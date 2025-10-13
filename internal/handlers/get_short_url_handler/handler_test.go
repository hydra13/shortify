package get_short_url_handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repo "github.com/hydra13/shortify/internal/repositories"
	inmemory_db "github.com/hydra13/shortify/internal/repositories/inmemory_db"
)

type GeneratorMock struct {
	wantInput string
	t         *testing.T
}

func (g GeneratorMock) GenerateShortID(in string) string {
	assert.Equal(g.t, g.wantInput, in)
	return "testing1"
}

func TestGetShortUrlHanderl_CreateHandler(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name       string
		url        string
		input      string
		repository repo.Repository
		generator  Generator
		want       want
	}{
		{
			name:       "positive test #1",
			url:        "/testing",
			input:      "https://ya.ru",
			repository: inmemory_db.New(),
			generator: GeneratorMock{
				wantInput: "https://ya.ru",
				t:         t,
			},
			want: want{
				code:        http.StatusCreated,
				response:    `http://localhost:8080/testing1`,
				contentType: "text/plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.url, strings.NewReader(tt.input))
			w := httptest.NewRecorder()
			handler := CreateHandler(tt.repository, tt.generator)

			handler.ServeHTTP(w, request)

			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, tt.want.response, string(resBody))
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
