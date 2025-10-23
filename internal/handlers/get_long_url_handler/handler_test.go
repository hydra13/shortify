package getlongurlhandler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type KeeperMock struct{}

func (k KeeperMock) Get(_ context.Context, shortURL string) (string, bool, error) {
	if shortURL == "testing1" {
		return "https://ya.ru", true, nil
	}

	if shortURL == "errorTst" {
		return "", false, errors.New("error")
	}

	return "", false, nil
}

func TestGetLongUrlHanderl_CreateHandler(t *testing.T) {
	keeperMock := &KeeperMock{}

	type want struct {
		code     int
		location string
	}

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "Success",
			url:  "/testing1",
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: `https://ya.ru`,
			},
		},
		{
			name: "Error - key length greater than in config",
			url:  "/testing1-testing-long-query",
			want: want{
				code:     http.StatusBadRequest,
				location: "",
			},
		},
		{
			name: "Error - key length less than in config",
			url:  "/t",
			want: want{
				code:     http.StatusBadRequest,
				location: "",
			},
		},
		{
			name: "Error - key not found",
			url:  "/testing2",
			want: want{
				code:     http.StatusNotFound,
				location: "",
			},
		},
		{
			name: "Error - during get value from repository",
			url:  "/errorTst",
			want: want{
				code:     http.StatusInternalServerError,
				location: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			handler := CreateHandler(keeperMock)

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
