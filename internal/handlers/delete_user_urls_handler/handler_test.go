package deleteuserurlshandler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hydra13/shortify/internal/handlers/delete_user_urls_handler/mocks"
	authContext "github.com/hydra13/shortify/internal/services/auth_context"
)

func TestGetShortUrlsBatchHanderl_CreateHandler(t *testing.T) {
	tests := []struct {
		name      string
		keeper    func(mc *minimock.Controller) UrlsKeeper
		userID    string
		isNewUser bool
		input     string
		wantCode  int
	}{
		{
			name: "Success",
			keeper: func(mc *minimock.Controller) UrlsKeeper {
				return mocks.NewUrlsKeeperMock(mc).
					DeleteAsyncMock.
					Expect(minimock.AnyContext, "user1", []string{"short1", "short2"}).Return()
			},
			userID:    "user1",
			isNewUser: false,
			input:     `["short1", "short2"]`,
			wantCode:  http.StatusAccepted,
		},
		{
			name:      "Error decode json",
			keeper:    func(mc *minimock.Controller) UrlsKeeper { return mocks.NewUrlsKeeperMock(mc) },
			userID:    "user1",
			isNewUser: false,
			input:     `test`,
			wantCode:  http.StatusBadRequest,
		},
		{
			name:      "Error by user",
			keeper:    func(mc *minimock.Controller) UrlsKeeper { return mocks.NewUrlsKeeperMock(mc) },
			isNewUser: true,
			input:     `["short1", "short2"]`,
			wantCode:  http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			keeper := tt.keeper(mc)

			handler := NewHandler(keeper, zerolog.Nop())

			ctx := authContext.CreateContextWithUserID(context.Background(), tt.userID, tt.isNewUser)

			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodDelete,
				"/api/user/urls",
				strings.NewReader(tt.input),
			)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.Handle(rr, req)

			assert.Equal(t, tt.wantCode, rr.Code)
		})
	}
}
