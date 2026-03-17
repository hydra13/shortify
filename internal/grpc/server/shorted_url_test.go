package grpcserver

import (
	"context"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/hydra13/shortify/api"
	"github.com/hydra13/shortify/internal/grpc/server/mocks"
	"github.com/hydra13/shortify/internal/models"
	"github.com/hydra13/shortify/internal/services/auth"
	authContext "github.com/hydra13/shortify/internal/services/auth_context"
)

func TestShortenURL(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*minimock.Controller) (*Server, *api.URLShortenRequest, context.Context)
		wantResult    string
		wantErr       bool
		wantErrorCode codes.Code
		wantErrorMsg  string
	}{
		{
			name: "success",
			setupMocks: func(mc *minimock.Controller) (*Server, *api.URLShortenRequest, context.Context) {
				shorter := mocks.NewShorterMock(mc).
					CreateMock.
					Expect(minimock.AnyContext, "https://ya.ru").
					Return("http://localhost:8080/abc123", nil)

				urlsKeeper := mocks.NewUrlsKeeperMock(mc)
				authService := auth.New()
				audit := mocks.NewAuditServiceMock(mc).
					PublishShortenEventMock.
					Expect("https://ya.ru", "").
					Return()

				server := NewServer(shorter, urlsKeeper, authService, audit, zerolog.Nop())
				req := &api.URLShortenRequest{Url: "https://ya.ru"}
				return server, req, context.Background()
			},
			wantResult: "http://localhost:8080/abc123",
			wantErr:    false,
		},
		{
			name: "empty url",
			setupMocks: func(mc *minimock.Controller) (*Server, *api.URLShortenRequest, context.Context) {
				server := NewServer(
					mocks.NewShorterMock(mc),
					mocks.NewUrlsKeeperMock(mc),
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)
				req := &api.URLShortenRequest{Url: ""}
				return server, req, context.Background()
			},
			wantErr:       true,
			wantErrorCode: codes.InvalidArgument,
			wantErrorMsg:  "url is required",
		},
		{
			name: "validation error",
			setupMocks: func(mc *minimock.Controller) (*Server, *api.URLShortenRequest, context.Context) {
				shorter := mocks.NewShorterMock(mc).
					CreateMock.
					Expect(minimock.AnyContext, "not-url").
					Return("", models.ErrValidation)

				server := NewServer(
					shorter,
					mocks.NewUrlsKeeperMock(mc),
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)
				req := &api.URLShortenRequest{Url: "not-url"}
				return server, req, context.Background()
			},
			wantErr:       true,
			wantErrorCode: codes.InvalidArgument,
			wantErrorMsg:  "invalid url",
		},
		{
			name: "conflict",
			setupMocks: func(mc *minimock.Controller) (*Server, *api.URLShortenRequest, context.Context) {
				shorter := mocks.NewShorterMock(mc).
					CreateMock.
					Expect(minimock.AnyContext, "https://ya.ru").
					Return("http://localhost:8080/existing", models.ErrConflict)

				server := NewServer(
					shorter,
					mocks.NewUrlsKeeperMock(mc),
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)
				req := &api.URLShortenRequest{Url: "https://ya.ru"}
				return server, req, context.Background()
			},
			wantResult: "http://localhost:8080/existing",
			wantErr:    false,
		},
		{
			name: "internal error",
			setupMocks: func(mc *minimock.Controller) (*Server, *api.URLShortenRequest, context.Context) {
				shorter := mocks.NewShorterMock(mc).
					CreateMock.
					Expect(minimock.AnyContext, "https://ya.ru").
					Return("", models.ErrInternal)

				server := NewServer(
					shorter,
					mocks.NewUrlsKeeperMock(mc),
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)
				req := &api.URLShortenRequest{Url: "https://ya.ru"}
				return server, req, context.Background()
			},
			wantErr:       true,
			wantErrorCode: codes.Internal,
			wantErrorMsg:  "internal server error",
		},
		{
			name: "with auth",
			setupMocks: func(mc *minimock.Controller) (*Server, *api.URLShortenRequest, context.Context) {
				userID := "test-user-123"

				shorter := mocks.NewShorterMock(mc).
					CreateMock.
					Inspect(func(ctx context.Context, long string) {
						extractedUserID, isNew := authContext.GetUserIDFromContext(ctx)
						assert.Equal(t, userID, extractedUserID, "userID should match")
						assert.False(t, isNew, "isNew should be false")
					}).
					Return("http://localhost:8080/abc123", nil)

				urlsKeeper := mocks.NewUrlsKeeperMock(mc)
				authService := auth.New()
				token, err := authService.BuildJWTString(userID)
				require.NoError(t, err)

				audit := mocks.NewAuditServiceMock(mc).
					PublishShortenEventMock.
					Expect("https://ya.ru", userID).
					Return()

				server := NewServer(shorter, urlsKeeper, authService, audit, zerolog.Nop())
				req := &api.URLShortenRequest{Url: "https://ya.ru"}

				ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					"authorization", "Bearer "+token,
				))

				return server, req, ctx
			},
			wantResult: "http://localhost:8080/abc123",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)
			server, req, ctx := tt.setupMocks(mc)

			resp, err := server.ShortenURL(ctx, req)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)

				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantErrorCode, st.Code())
				if tt.wantErrorMsg != "" {
					assert.Equal(t, tt.wantErrorMsg, st.Message())
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, resp.Result)
			}
		})
	}
}
