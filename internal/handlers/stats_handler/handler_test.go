package statshandler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/hydra13/shortify/internal/handlers/stats_handler/mocks"
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

func TestStatsHandler_Handle(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name              string
		repository        func(mc *minimock.Controller) Repository
		want              want
		skipCheckResponse bool
	}{
		{
			name: "success with positive values",
			repository: func(mc *minimock.Controller) Repository {
				return mocks.NewRepositoryMock(mc).
					GetStatsMock.
					Expect(minimock.AnyContext).
					Return(100, 50, nil)
			},
			want: want{
				code:        http.StatusOK,
				response:    `{"urls":100,"users":50}`,
				contentType: "application/json",
			},
			skipCheckResponse: false,
		},
		{
			name: "success with zero values",
			repository: func(mc *minimock.Controller) Repository {
				return mocks.NewRepositoryMock(mc).
					GetStatsMock.
					Expect(minimock.AnyContext).
					Return(0, 0, nil)
			},
			want: want{
				code:        http.StatusOK,
				response:    `{"urls":0,"users":0}`,
				contentType: "application/json",
			},
			skipCheckResponse: false,
		},
		{
			name: "success with urls only",
			repository: func(mc *minimock.Controller) Repository {
				return mocks.NewRepositoryMock(mc).
					GetStatsMock.
					Expect(minimock.AnyContext).
					Return(25, 0, nil)
			},
			want: want{
				code:        http.StatusOK,
				response:    `{"urls":25,"users":0}`,
				contentType: "application/json",
			},
			skipCheckResponse: false,
		},
		{
			name: "success with users only",
			repository: func(mc *minimock.Controller) Repository {
				return mocks.NewRepositoryMock(mc).
					GetStatsMock.
					Expect(minimock.AnyContext).
					Return(0, 10, nil)
			},
			want: want{
				code:        http.StatusOK,
				response:    `{"urls":0,"users":10}`,
				contentType: "application/json",
			},
			skipCheckResponse: false,
		},
		{
			name: "success with large values",
			repository: func(mc *minimock.Controller) Repository {
				return mocks.NewRepositoryMock(mc).
					GetStatsMock.
					Expect(minimock.AnyContext).
					Return(1000000, 500000, nil)
			},
			want: want{
				code:        http.StatusOK,
				response:    `{"urls":1000000,"users":500000}`,
				contentType: "application/json",
			},
			skipCheckResponse: false,
		},
		{
			name: "success with equal urls and users",
			repository: func(mc *minimock.Controller) Repository {
				return mocks.NewRepositoryMock(mc).
					GetStatsMock.
					Expect(minimock.AnyContext).
					Return(42, 42, nil)
			},
			want: want{
				code:        http.StatusOK,
				response:    `{"urls":42,"users":42}`,
				contentType: "application/json",
			},
			skipCheckResponse: false,
		},
		{
			name: "error from repository",
			repository: func(mc *minimock.Controller) Repository {
				return mocks.NewRepositoryMock(mc).
					GetStatsMock.
					Expect(minimock.AnyContext).
					Return(0, 0, errors.New("database connection failed"))
			},
			want: want{
				code: http.StatusInternalServerError,
			},
			skipCheckResponse: true,
		},
		{
			name: "error with context timeout",
			repository: func(mc *minimock.Controller) Repository {
				return mocks.NewRepositoryMock(mc).
					GetStatsMock.
					Expect(minimock.AnyContext).
					Return(0, 0, errors.New("context deadline exceeded"))
			},
			want: want{
				code: http.StatusInternalServerError,
			},
			skipCheckResponse: true,
		},
		{
			name: "success with single url and user",
			repository: func(mc *minimock.Controller) Repository {
				return mocks.NewRepositoryMock(mc).
					GetStatsMock.
					Expect(minimock.AnyContext).
					Return(1, 1, nil)
			},
			want: want{
				code:        http.StatusOK,
				response:    `{"urls":1,"users":1}`,
				contentType: "application/json",
			},
			skipCheckResponse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)

			repo := tt.repository(mc)

			handler := NewHandler(repo, zerolog.Nop())

			srv := httptest.NewServer(&wrapper{h: handler})
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/stats", nil)
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

			assert.Equal(t, tt.want.response, strings.TrimSpace(string(body)))
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}
