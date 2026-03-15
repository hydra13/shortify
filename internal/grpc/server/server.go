//go:generate minimock -i .Shorter,.UrlsKeeper,.AuditService -o mocks -s _mock.go -g
package grpcserver

import (
	"context"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/hydra13/shortify/api"
	"github.com/hydra13/shortify/internal/services/auth"
)

type Shorter interface {
	Create(ctx context.Context, long string) (string, error)
	CreateShortURL(shortID string) string
}

type UrlsKeeper interface {
	Get(ctx context.Context, shortURL string) (url string, found bool, err error)
	GetAllByUser(ctx context.Context, userID string) (urls map[string]string, err error)
}

type AuditService interface {
	PublishShortenEvent(longURL, userID string)
	PublishFollowEvent(longURL, userID string)
}

type Server struct {
	api.UnimplementedShortenerServiceServer
	shorter    Shorter
	urlsKeeper UrlsKeeper
	auth       *auth.AuthService
	audit      AuditService
	log        zerolog.Logger
}

func NewServer(
	shorter Shorter,
	urlsKeeper UrlsKeeper,
	authService *auth.AuthService,
	auditService AuditService,
	log zerolog.Logger,
) *Server {
	return &Server{
		shorter:    shorter,
		urlsKeeper: urlsKeeper,
		auth:       authService,
		audit:      auditService,
		log:        log,
	}
}

func (s *Server) Register(grpcServer *grpc.Server) {
	api.RegisterShortenerServiceServer(grpcServer, s)
}

func (s *Server) getUserIDFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata not found")
	}

	authValues := md.Get("authorization")
	if len(authValues) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization header not found")
	}

	token := authValues[0]
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	userID, err := s.auth.GetUserIDFromToken(token)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid token")
	}

	return userID, nil
}
