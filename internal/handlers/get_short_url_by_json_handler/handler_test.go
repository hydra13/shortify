package getshorturlbyjsonhandler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hydra13/shortify/internal/handlers/get_short_url_by_json_handler/mocks"
	"github.com/hydra13/shortify/internal/models"
)

type wrapper struct {
	h *Handler
}

func (wr *wrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wr.h.Handle(w, r)
}

func TestGetShortUrlHanderl_CreateHandler(t *testing.T) {
	var buf bytes.Buffer
	log := zerolog.New(&buf)
	zerolog.SetGlobalLevel(zerolog.Disabled)
	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name    string
		shorter func(mc *minimock.Controller) Shorter
		auth    func(mc *minimock.Controller) AuthService
		url     string
		input   string
		want    want
		wantErr bool
	}{
		{
			name: "Success",
			shorter: func(mc *minimock.Controller) Shorter {
				return mocks.NewShorterMock(mc).
					CreateMock.
					Expect(minimock.AnyContext, "https://ya.ru").
					Return("http://localhost:8080/testing1", nil)
			},
			auth: func(mc *minimock.Controller) AuthService {
				return nil // mocks.NewAuthServiceMock(mc)
			},
			url:   "/api/shorten",
			input: `{"url":"https://ya.ru"}`,
			want: want{
				code:        http.StatusCreated,
				response:    "{\"result\":\"http://localhost:8080/testing1\"}\n",
				contentType: "application/json",
			},
		},
		{
			name:    "Error when body is empty",
			shorter: func(mc *minimock.Controller) Shorter { return mocks.NewShorterMock(mc) },
			auth:    func(mc *minimock.Controller) AuthService { return nil }, // mocks.NewAuthServiceMock(mc) },
			input:   "",
			url:     "/api/shorten",
			want: want{
				code: http.StatusBadRequest,
			},
			wantErr: true,
		},
		{
			name: "Error when input is not url",
			shorter: func(mc *minimock.Controller) Shorter {
				return mocks.NewShorterMock(mc).
					CreateMock.
					Expect(minimock.AnyContext, "not-url").
					Return("", models.ErrValidation)
			},
			auth:  func(mc *minimock.Controller) AuthService { return nil }, // mocks.NewAuthServiceMock(mc) },
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
			t.Parallel()

			mc := minimock.NewController(t)
			shorter := tt.shorter(mc)
			auth := tt.auth(mc)

			handler := NewHandler(shorter, auth, log)

			srv := httptest.NewServer(&wrapper{h: handler})
			defer srv.Close()

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
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
