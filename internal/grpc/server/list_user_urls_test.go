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
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/hydra13/shortify/api"
	"github.com/hydra13/shortify/internal/grpc/server/mocks"
	"github.com/hydra13/shortify/internal/services/auth"
)

func TestListUserURLs(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*minimock.Controller) (*Server, context.Context)
		wantURLs      int
		wantURLData   []*api.URLData
		wantErr       bool
		wantErrorCode codes.Code
		wantErrorMsg  string
	}{
		{
			name: "success",
			setupMocks: func(mc *minimock.Controller) (*Server, context.Context) {
				userID := "test-user-123"
				token, err := auth.New().BuildJWTString(userID)
				require.NoError(t, err)

				urls := map[string]string{
					"abc123": "https://ya.ru",
				}

				urlsKeeper := mocks.NewUrlsKeeperMock(mc).
					GetAllByUserMock.
					Expect(minimock.AnyContext, userID).
					Return(urls, nil)

				shorter := mocks.NewShorterMock(mc).
					CreateShortURLMock.
					Expect("abc123").
					Return("http://localhost:8080/abc123")

				server := NewServer(
					shorter,
					urlsKeeper,
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)

				ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					"authorization", token,
				))

				return server, ctx
			},
			wantURLs: 1,
			wantURLData: []*api.URLData{
				{
					ShortUrl:    "http://localhost:8080/abc123",
					OriginalUrl: "https://ya.ru",
				},
			},
			wantErr: false,
		},
		{
			name: "empty list",
			setupMocks: func(mc *minimock.Controller) (*Server, context.Context) {
				userID := "test-user-123"
				token, err := auth.New().BuildJWTString(userID)
				require.NoError(t, err)

				urlsKeeper := mocks.NewUrlsKeeperMock(mc).
					GetAllByUserMock.
					Expect(minimock.AnyContext, userID).
					Return(map[string]string{}, nil)

				server := NewServer(
					mocks.NewShorterMock(mc),
					urlsKeeper,
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)

				ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					"authorization", token,
				))

				return server, ctx
			},
			wantURLs: 0,
			wantErr:  false,
		},
		{
			name: "unauthorized",
			setupMocks: func(mc *minimock.Controller) (*Server, context.Context) {
				server := NewServer(
					mocks.NewShorterMock(mc),
					mocks.NewUrlsKeeperMock(mc),
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)

				return server, context.Background()
			},
			wantErr:       true,
			wantErrorCode: codes.Unauthenticated,
		},
		{
			name: "invalid token",
			setupMocks: func(mc *minimock.Controller) (*Server, context.Context) {
				server := NewServer(
					mocks.NewShorterMock(mc),
					mocks.NewUrlsKeeperMock(mc),
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)

				ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					"authorization", "invalid-token",
				))

				return server, ctx
			},
			wantErr:       true,
			wantErrorCode: codes.Unauthenticated,
		},
		{
			name: "internal error",
			setupMocks: func(mc *minimock.Controller) (*Server, context.Context) {
				userID := "test-user-123"
				token, err := auth.New().BuildJWTString(userID)
				require.NoError(t, err)

				urlsKeeper := mocks.NewUrlsKeeperMock(mc).
					GetAllByUserMock.
					Expect(minimock.AnyContext, userID).
					Return(nil, assert.AnError)

				server := NewServer(
					mocks.NewShorterMock(mc),
					urlsKeeper,
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)

				ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
					"authorization", token,
				))

				return server, ctx
			},
			wantErr:       true,
			wantErrorCode: codes.Internal,
			wantErrorMsg:  "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)
			server, ctx := tt.setupMocks(mc)

			req := &emptypb.Empty{}
			resp, err := server.ListUserURLs(ctx, req)

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
				assert.Len(t, resp.Url, tt.wantURLs)

				if tt.wantURLData != nil {
					for i, expectedURL := range tt.wantURLData {
						assert.Equal(t, expectedURL.ShortUrl, resp.Url[i].ShortUrl)
						assert.Equal(t, expectedURL.OriginalUrl, resp.Url[i].OriginalUrl)
					}
				}
			}
		})
	}
}
