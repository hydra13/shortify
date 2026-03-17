package grpcserver

import (
	"context"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hydra13/shortify/api"
	"github.com/hydra13/shortify/internal/grpc/server/mocks"
	"github.com/hydra13/shortify/internal/models"
	"github.com/hydra13/shortify/internal/services/auth"
)

func TestExpandURL(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*minimock.Controller) (*Server, *api.URLExpandRequest, context.Context)
		wantResult    string
		wantErr       bool
		wantErrorCode codes.Code
		wantErrorMsg  string
	}{
		{
			name: "success",
			setupMocks: func(mc *minimock.Controller) (*Server, *api.URLExpandRequest, context.Context) {
				urlsKeeper := mocks.NewUrlsKeeperMock(mc).
					GetMock.
					Expect(minimock.AnyContext, "abc123").
					Return("https://ya.ru", true, nil)

				audit := mocks.NewAuditServiceMock(mc).
					PublishFollowEventMock.
					Expect("https://ya.ru", "").
					Return()

				server := NewServer(
					mocks.NewShorterMock(mc),
					urlsKeeper,
					auth.New(),
					audit,
					zerolog.Nop(),
				)
				req := &api.URLExpandRequest{Id: "abc123"}
				return server, req, context.Background()
			},
			wantResult: "https://ya.ru",
			wantErr:    false,
		},
		{
			name: "empty id",
			setupMocks: func(mc *minimock.Controller) (*Server, *api.URLExpandRequest, context.Context) {
				server := NewServer(
					mocks.NewShorterMock(mc),
					mocks.NewUrlsKeeperMock(mc),
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)
				req := &api.URLExpandRequest{Id: ""}
				return server, req, context.Background()
			},
			wantErr:       true,
			wantErrorCode: codes.InvalidArgument,
			wantErrorMsg:  "id is required",
		},
		{
			name: "not found",
			setupMocks: func(mc *minimock.Controller) (*Server, *api.URLExpandRequest, context.Context) {
				urlsKeeper := mocks.NewUrlsKeeperMock(mc).
					GetMock.
					Expect(minimock.AnyContext, "nonexistent").
					Return("", false, nil)

				server := NewServer(
					mocks.NewShorterMock(mc),
					urlsKeeper,
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)
				req := &api.URLExpandRequest{Id: "nonexistent"}
				return server, req, context.Background()
			},
			wantErr:       true,
			wantErrorCode: codes.NotFound,
			wantErrorMsg:  "url not found",
		},
		{
			name: "deleted",
			setupMocks: func(mc *minimock.Controller) (*Server, *api.URLExpandRequest, context.Context) {
				urlsKeeper := mocks.NewUrlsKeeperMock(mc).
					GetMock.
					Expect(minimock.AnyContext, "deleted123").
					Return("", false, models.ErrURLIsDeleted)

				server := NewServer(
					mocks.NewShorterMock(mc),
					urlsKeeper,
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)
				req := &api.URLExpandRequest{Id: "deleted123"}
				return server, req, context.Background()
			},
			wantErr:       true,
			wantErrorCode: codes.NotFound,
			wantErrorMsg:  "url is deleted",
		},
		{
			name: "internal error",
			setupMocks: func(mc *minimock.Controller) (*Server, *api.URLExpandRequest, context.Context) {
				urlsKeeper := mocks.NewUrlsKeeperMock(mc).
					GetMock.
					Expect(minimock.AnyContext, "error123").
					Return("", false, assert.AnError)

				server := NewServer(
					mocks.NewShorterMock(mc),
					urlsKeeper,
					auth.New(),
					mocks.NewAuditServiceMock(mc),
					zerolog.Nop(),
				)
				req := &api.URLExpandRequest{Id: "error123"}
				return server, req, context.Background()
			},
			wantErr:       true,
			wantErrorCode: codes.Internal,
			wantErrorMsg:  "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)
			server, req, ctx := tt.setupMocks(mc)

			resp, err := server.ExpandURL(ctx, req)

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
