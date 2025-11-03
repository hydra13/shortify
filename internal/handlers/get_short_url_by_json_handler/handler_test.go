package getshorturlbyjsonhandler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hydra13/shortify/internal/models"
)

type ShorterMock struct{}

func (s ShorterMock) Create(_ context.Context, long string) (string, error) {
	if long == "https://ya.ru" {
		return "http://localhost:8080/testing1", nil
	}

	return "", models.ErrValidation
}

func TestGetShortUrlHanderl_CreateHandler(t *testing.T) {
	shorter := &ShorterMock{}

	handler := CreateHandler(shorter, "http://localhost:8080")
	srv := httptest.NewServer(handler)

	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name    string
		url     string
		input   string
		want    want
		wantErr bool
	}{
		{
			name:  "Success",
			url:   "/api/shorten",
			input: `{"url":"https://ya.ru"}`,
			want: want{
				code:        http.StatusCreated,
				response:    "{\"result\":\"http://localhost:8080/testing1\"}\n",
				contentType: "application/json",
			},
		},
		{
			name:  "Error when body is empty",
			input: "",
			url:   "/api/shorten",
			want: want{
				code: http.StatusBadRequest,
			},
			wantErr: true,
		},
		{
			name:  "Error when input is not url",
			input: `{"url":"not-url"}`,
			url:   "/api/shorten",
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
