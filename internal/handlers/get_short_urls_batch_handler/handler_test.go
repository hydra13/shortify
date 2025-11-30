package getshorturlsbatchhandler

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

	"github.com/hydra13/shortify/internal/handlers/get_short_urls_batch_handler/mocks"
	"github.com/hydra13/shortify/internal/models"
)

func TestGetShortUrlsBatchHanderl_CreateHandler(t *testing.T) {
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
		url     string
		input   string
		want    want
		wantErr bool
	}{
		{
			name: "Success",
			shorter: func(mc *minimock.Controller) Shorter {
				return mocks.NewShorterMock(mc).
					CreateBatchMock.
					Expect(minimock.AnyContext, map[string]string{
						"correlation_id1": "https://ya.ru",
					}).
					Return(map[string]string{
						"correlation_id1": "http://localhost:8080/testing1",
					}, nil)
			},
			url:   "/api/shorten/batch",
			input: `[{"correlation_id":"correlation_id1", "original_url":"https://ya.ru"}]`,
			want: want{
				code:        http.StatusCreated,
				response:    `[{"correlation_id":"correlation_id1","short_url":"http://localhost:8080/testing1"}]`,
				contentType: "application/json",
			},
		},
		{
			name: "Error when body is empty",
			shorter: func(mc *minimock.Controller) Shorter {
				return mocks.NewShorterMock(mc)
			},
			input: "",
			url:   "/api/shorten/batch",
			want: want{
				code: http.StatusBadRequest,
			},
			wantErr: true,
		},
		{
			name: "Error when input is not url",
			shorter: func(mc *minimock.Controller) Shorter {
				return mocks.NewShorterMock(mc).
					CreateBatchMock.
					Expect(minimock.AnyContext, map[string]string{
						"correlation_id1": "not-url",
					}).
					Return(nil, models.ErrValidation)
			},
			input: `[{"correlation_id":"correlation_id1", "original_url":"not-url"}]`,
			url:   "/api/shorten/batch",
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

			handler := CreateHandler(shorter, log)
			srv := httptest.NewServer(handler)
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
			assert.Equal(t, tt.want.response, strings.Trim(string(body), "\n"))
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
