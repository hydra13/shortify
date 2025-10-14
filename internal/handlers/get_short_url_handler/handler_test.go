package getshorturlhandler

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
	generatorMock := GeneratorMock{
		wantInput: "https://ya.ru",
		t:         t,
	}

	dbMock := inmemory_db.New()

	handler := CreateHandler(dbMock, generatorMock, "http://localhost:8080")
	srv := httptest.NewServer(handler)

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
		wantErr    bool
	}{
		{
			name:       "Success",
			url:        "/testing1",
			input:      "https://ya.ru",
			repository: dbMock,
			generator:  generatorMock,
			want: want{
				code:        http.StatusCreated,
				response:    `http://localhost:8080/testing1`,
				contentType: "text/plain",
			},
		},
		{
			name:       "Error when body is empty",
			input:      "",
			url:        "/",
			repository: dbMock,
			generator:  generatorMock,
			want: want{
				code: http.StatusBadRequest,
			},
			wantErr: true,
		},
		{
			name:       "Error when input is not url",
			input:      "not-url",
			url:        "/",
			repository: dbMock,
			generator:  generatorMock,
			want: want{
				code: http.StatusBadRequest,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, srv.URL+tt.url, strings.NewReader(tt.input))
			require.NoError(t, err)

			resp, err := srv.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode)

			if tt.wantErr {
				return
			}

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.response, string(body))
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}
