package getlongurlhandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	repo "github.com/hydra13/shortify/internal/repositories"
	inmemory_db "github.com/hydra13/shortify/internal/repositories/inmemory_db"
)

func TestGetLongUrlHanderl_CreateHandler(t *testing.T) {
	repository := inmemory_db.New()
	repository.Add("testing1", "https://ya.ru")

	repositoryErr := &repo.RepositoryErrorMock{}

	type want struct {
		code     int
		location string
	}

	tests := []struct {
		name       string
		url        string
		repository repo.Repository
		want       want
	}{
		{
			name:       "Success",
			url:        "/testing1",
			repository: repository,
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: `https://ya.ru`,
			},
		},
		{
			name:       "Error - key length greater than in config",
			url:        "/testing1-testing-long-query",
			repository: repository,
			want: want{
				code:     http.StatusBadRequest,
				location: "",
			},
		},
		{
			name:       "Error - key length less than in config",
			url:        "/t",
			repository: repository,
			want: want{
				code:     http.StatusBadRequest,
				location: "",
			},
		},
		{
			name:       "Error - key not found",
			url:        "/testing2",
			repository: repository,
			want: want{
				code:     http.StatusBadRequest,
				location: "",
			},
		},
		{
			name:       "Error - during get value from repository",
			url:        "/testing1",
			repository: repositoryErr,
			want: want{
				code:     http.StatusBadRequest,
				location: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			handler := CreateHandler(tt.repository)

			handler.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)

			if tt.want.location != "" {
				location := res.Header.Get("Location")
				assert.Equal(t, tt.want.location, location)
			}
		})
	}
}
