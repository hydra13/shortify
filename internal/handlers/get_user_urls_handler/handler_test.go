package getuserurlshandler

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/hydra13/shortify/internal/handlers/get_user_urls_handler/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wrapper struct {
	h *Handler
}

func (wr *wrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wr.h.Handle(w, r)
}

func TestGetUserUrlsHandler_CreateHandler(t *testing.T) {
	var buf bytes.Buffer
	log := zerolog.New(&buf)
	zerolog.SetGlobalLevel(zerolog.Disabled)
	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name              string
		userID            string
		urlsKeeper        func(mc *minimock.Controller) UrlsKeeper
		shorter           func(mc *minimock.Controller) Shorter
		want              want
		skipCheckResponse bool
	}{
		{
			name:   "Success",
			userID: "user1",
			urlsKeeper: func(mc *minimock.Controller) UrlsKeeper {
				return mocks.NewUrlsKeeperMock(mc).
					GetAllByUserMock.
					Return(map[string]string{
						"shortURL1": "https://ya.ru",
					}, nil)
			},
			shorter: func(mc *minimock.Controller) Shorter {
				return mocks.NewShorterMock(mc).
					CreateShortURLMock.
					Expect("shortURL1").
					Return("http://localhost:8080/shortURL1")
			},
			want: want{
				code:        http.StatusOK,
				response:    `[{"short_url":"http://localhost:8080/shortURL1","original_url":"https://ya.ru"}]`,
				contentType: "application/json",
			},
			skipCheckResponse: false,
		},
		{
			name:   "Error by UrlsKeeper",
			userID: "user1",
			urlsKeeper: func(mc *minimock.Controller) UrlsKeeper {
				return mocks.NewUrlsKeeperMock(mc).
					GetAllByUserMock.
					Return(nil, errors.New("some error"))
			},
			shorter: func(mc *minimock.Controller) Shorter { return mocks.NewShorterMock(mc) },
			want: want{
				code: http.StatusInternalServerError,
			},
			skipCheckResponse: true,
		},
		{
			name:   "Success",
			userID: "user1",
			urlsKeeper: func(mc *minimock.Controller) UrlsKeeper {
				return mocks.NewUrlsKeeperMock(mc).
					GetAllByUserMock.
					Return(map[string]string{}, nil)
			},
			shorter: func(mc *minimock.Controller) Shorter { return mocks.NewShorterMock(mc) },
			want: want{
				code: http.StatusNoContent,
			},
			skipCheckResponse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)

			urlsKeeper := tt.urlsKeeper(mc)
			shorter := tt.shorter(mc)

			handler := NewHandler(urlsKeeper, shorter, log)

			srv := httptest.NewServer(&wrapper{h: handler})
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/user/urls", nil)
			require.NoError(t, err)

			resp, err := srv.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode)

			if tt.skipCheckResponse {
				return
			}

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.response, strings.Trim(string(body), "\n"))
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
