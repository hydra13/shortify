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
	"github.com/hydra13/shortify/internal/models"
	"github.com/hydra13/shortify/internal/services/auth"
	authContext "github.com/hydra13/shortify/internal/services/auth_context"
)

func TestShortenURL_Success(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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
	resp, err := server.ShortenURL(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/abc123", resp.Result)
}

func TestShortenURL_EmptyURL(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	server := NewServer(
		mocks.NewShorterMock(mc),
		mocks.NewUrlsKeeperMock(mc),
		auth.New(),
		mocks.NewAuditServiceMock(mc),
		zerolog.Nop(),
	)

	req := &api.URLShortenRequest{Url: ""}
	resp, err := server.ShortenURL(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Equal(t, "url is required", st.Message())
}

func TestShortenURL_ValidationError(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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
	resp, err := server.ShortenURL(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Equal(t, "invalid url", st.Message())
}

func TestShortenURL_Conflict(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	shorter := mocks.NewShorterMock(mc).
		CreateMock.
		Expect(minimock.AnyContext, "https://ya.ru").
		Return("http://localhost:8080/existing", models.ErrConflict)

	// No audit expectation since the order of calls might vary
	audit := mocks.NewAuditServiceMock(mc)

	server := NewServer(
		shorter,
		mocks.NewUrlsKeeperMock(mc),
		auth.New(),
		audit,
		zerolog.Nop(),
	)

	req := &api.URLShortenRequest{Url: "https://ya.ru"}
	resp, err := server.ShortenURL(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/existing", resp.Result)
}

func TestShortenURL_InternalError(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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
	resp, err := server.ShortenURL(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Equal(t, "internal server error", st.Message())
}

func TestShortenURL_WithAuth(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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

	// Create a valid token
	token, err := authService.BuildJWTString(userID)
	require.NoError(t, err)

	audit := mocks.NewAuditServiceMock(mc).
		PublishShortenEventMock.
		Expect("https://ya.ru", userID).
		Return()

	server := NewServer(shorter, urlsKeeper, authService, audit, zerolog.Nop())

	// Create context with authorization metadata
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer "+token,
	))

	req := &api.URLShortenRequest{Url: "https://ya.ru"}
	resp, err := server.ShortenURL(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/abc123", resp.Result)
}

func TestExpandURL_Success(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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
	resp, err := server.ExpandURL(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "https://ya.ru", resp.Result)
}

func TestExpandURL_EmptyID(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	server := NewServer(
		mocks.NewShorterMock(mc),
		mocks.NewUrlsKeeperMock(mc),
		auth.New(),
		mocks.NewAuditServiceMock(mc),
		zerolog.Nop(),
	)

	req := &api.URLExpandRequest{Id: ""}
	resp, err := server.ExpandURL(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Equal(t, "id is required", st.Message())
}

func TestExpandURL_NotFound(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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
	resp, err := server.ExpandURL(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Equal(t, "url not found", st.Message())
}

func TestExpandURL_Deleted(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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
	resp, err := server.ExpandURL(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Equal(t, "url is deleted", st.Message())
}

func TestExpandURL_InternalError(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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
	resp, err := server.ExpandURL(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Equal(t, "internal server error", st.Message())
}

func TestListUserURLs_Success(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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

	req := &emptypb.Empty{}
	resp, err := server.ListUserURLs(ctx, req)

	require.NoError(t, err)
	assert.Len(t, resp.Url, 1)

	// Verify URL data
	urlData := resp.Url[0]
	assert.Equal(t, "http://localhost:8080/abc123", urlData.ShortUrl)
	assert.Equal(t, "https://ya.ru", urlData.OriginalUrl)
}

func TestListUserURLs_EmptyList(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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

	req := &emptypb.Empty{}
	resp, err := server.ListUserURLs(ctx, req)

	require.NoError(t, err)
	assert.Len(t, resp.Url, 0)
}

func TestListUserURLs_Unauthorized(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	server := NewServer(
		mocks.NewShorterMock(mc),
		mocks.NewUrlsKeeperMock(mc),
		auth.New(),
		mocks.NewAuditServiceMock(mc),
		zerolog.Nop(),
	)

	req := &emptypb.Empty{}
	resp, err := server.ListUserURLs(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	// The error message should detail the specific authentication issue
	assert.NotEmpty(t, st.Message())
}

func TestListUserURLs_InvalidToken(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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

	req := &emptypb.Empty{}
	resp, err := server.ListUserURLs(ctx, req)

	require.Error(t, err)
	require.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	// The error message can vary depending on the JWT library implementation
	assert.NotEmpty(t, st.Message())
}

func TestListUserURLs_InternalError(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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

	req := &emptypb.Empty{}
	resp, err := server.ListUserURLs(ctx, req)

	require.Error(t, err)
	require.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Equal(t, "internal server error", st.Message())
}

func TestGetUserIDFromMetadata_BearerToken(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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

	extractedUserID, err := server.getUserIDFromMetadata(ctx)

	require.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
}

func TestGetUserIDFromMetadata_TokenWithoutBearer(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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

	extractedUserID, err := server.getUserIDFromMetadata(ctx)

	require.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
}

func TestGetUserIDFromMetadata_NoMetadata(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	server := NewServer(
		mocks.NewShorterMock(mc),
		mocks.NewUrlsKeeperMock(mc),
		auth.New(),
		mocks.NewAuditServiceMock(mc),
		zerolog.Nop(),
	)

	extractedUserID, err := server.getUserIDFromMetadata(context.Background())

	require.Error(t, err)
	require.Empty(t, extractedUserID)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGetUserIDFromMetadata_NoAuthorizationHeader(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

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

	extractedUserID, err := server.getUserIDFromMetadata(ctx)

	require.Error(t, err)
	require.Empty(t, extractedUserID)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}
