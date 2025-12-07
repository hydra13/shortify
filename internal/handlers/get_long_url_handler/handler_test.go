package getlongurlhandler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/hydra13/shortify/internal/handlers/get_long_url_handler/mocks"
)

func TestGetLongUrlHanderl_CreateHandler(t *testing.T) {
	var buf bytes.Buffer
	log := zerolog.New(&buf)
	zerolog.SetGlobalLevel(zerolog.Disabled)
	type want struct {
		code     int
		location string
	}

	tests := []struct {
		name   string
		keeper func(mc *minimock.Controller) UrlsKeeper
		url    string
		want   want
	}{
		{
			name: "Success",
			keeper: func(mc *minimock.Controller) UrlsKeeper {
				return mocks.NewUrlsKeeperMock(mc).GetMock.
					Expect(minimock.AnyContext, "testing1").
					Return("https://ya.ru", true, nil)
			},
			url: "/testing1",
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: `https://ya.ru`,
			},
		},
		{
			name:   "Error - key length greater than in config",
			keeper: func(mc *minimock.Controller) UrlsKeeper { return mocks.NewUrlsKeeperMock(mc) },
			url:    "/testing1-testing-long-query",
			want: want{
				code:     http.StatusBadRequest,
				location: "",
			},
		},
		{
			name:   "Error - key length less than in config",
			keeper: func(mc *minimock.Controller) UrlsKeeper { return mocks.NewUrlsKeeperMock(mc) },
			url:    "/t",
			want: want{
				code:     http.StatusBadRequest,
				location: "",
			},
		},
		{
			name: "Error - key not found",
			keeper: func(mc *minimock.Controller) UrlsKeeper {
				return mocks.NewUrlsKeeperMock(mc).GetMock.
					Expect(minimock.AnyContext, "testing2").
					Return("", false, nil)
			},
			url: "/testing2",
			want: want{
				code:     http.StatusNotFound,
				location: "",
			},
		},
		{
			name: "Error - during get value from repository",
			keeper: func(mc *minimock.Controller) UrlsKeeper {
				return mocks.NewUrlsKeeperMock(mc).GetMock.
					Expect(minimock.AnyContext, "errorTst").
					Return("", false, errors.New("some error"))
			},
			url: "/errorTst",
			want: want{
				code:     http.StatusInternalServerError,
				location: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mc := minimock.NewController(t)
			keeper := tt.keeper(mc)

			request := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			handler := NewHandler(keeper, log)

			handler.Handle(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)

			if tt.want.location != "" {
				location := res.Header.Get("Location")
				assert.Equal(t, tt.want.location, location)
			}
		})
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
