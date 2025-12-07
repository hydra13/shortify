package pinghandler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hydra13/shortify/internal/handlers/ping_handler/mocks"
)

type wrapper struct {
	h *Handler
}

func (wr *wrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wr.h.Handle(w, r)
}

func TestPingHanderl_CreateHandler(t *testing.T) {
	var buf bytes.Buffer
	log := zerolog.New(&buf)
	zerolog.SetGlobalLevel(zerolog.Disabled)

	tests := []struct {
		name       string
		db         func(mc *minimock.Controller) DB
		statusCode int
	}{
		{
			name: "success",
			db: func(mc *minimock.Controller) DB {
				return mocks.NewDBMock(mc).
					PingContextMock.
					Expect(minimock.AnyContext).
					Return(nil)
			},
			statusCode: http.StatusOK,
		},
		{
			name: "error by connection db",
			db: func(mc *minimock.Controller) DB {
				return mocks.NewDBMock(mc).
					PingContextMock.
					Expect(minimock.AnyContext).
					Return(errors.New("some error"))
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "error by nil db instance",
			db: func(_ *minimock.Controller) DB {
				return nil
			},
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)

			db := tt.db(mc)

			handler := NewHandler(db, log)

			srv := httptest.NewServer(&wrapper{h: handler})
			defer srv.Close()

			resp, err := http.Get(srv.URL)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.statusCode, resp.StatusCode)
		})
	}
}
