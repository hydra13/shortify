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

	"github.com/hydra13/shortify/internal/grpc/server/mocks"
	"github.com/hydra13/shortify/internal/services/auth"
)

func TestGetUserIDFromMetadata(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func(*minimock.Controller) (*Server, context.Context)
		wantUserID    string
		wantErr       bool
		wantErrorCode codes.Code
	}{
		{
			name: "bearer token",
			setupContext: func(mc *minimock.Controller) (*Server, context.Context) {
				authService := auth.New()
				userID := "test-user-123"
				token, err := authService.BuildJWTString(userID)
				require.NoError(t, err)

				server := NewServer(
					mocks.NewShorterMock(mc),
					mocks.NewUrlsKeeperMock(mc),
					authService,
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)

				ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					"authorization", "Bearer "+token,
				))

				return server, ctx
			},
			wantUserID: "test-user-123",
			wantErr:    false,
		},
		{
			name: "token without bearer",
			setupContext: func(mc *minimock.Controller) (*Server, context.Context) {
				authService := auth.New()
				userID := "test-user-123"
				token, err := authService.BuildJWTString(userID)
				require.NoError(t, err)

				server := NewServer(
					mocks.NewShorterMock(mc),
					mocks.NewUrlsKeeperMock(mc),
					authService,
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)

				ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					"authorization", token,
				))

				return server, ctx
			},
			wantUserID: "test-user-123",
			wantErr:    false,
		},
		{
			name: "no metadata",
			setupContext: func(mc *minimock.Controller) (*Server, context.Context) {
				server := NewServer(
					mocks.NewShorterMock(mc),
					mocks.NewUrlsKeeperMock(mc),
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)

				ctx := context.Background()
				return server, ctx
			},
			wantUserID:    "",
			wantErr:       true,
			wantErrorCode: codes.Unauthenticated,
		},
		{
			name: "no authorization header",
			setupContext: func(mc *minimock.Controller) (*Server, context.Context) {
				server := NewServer(
					mocks.NewShorterMock(mc),
					mocks.NewUrlsKeeperMock(mc),
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)

				ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					"other-header", "value",
				))

				return server, ctx
			},
			wantUserID:    "",
			wantErr:       true,
			wantErrorCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)
			server, ctx := tt.setupContext(mc)

			extractedUserID, err := server.getUserIDFromMetadata(ctx)

			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, extractedUserID)

				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantErrorCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantUserID, extractedUserID)
			}
		})
	}
}
