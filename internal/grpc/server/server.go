//go:generate minimock -i .Shorter,.UrlsKeeper,.AuditService -o mocks -s _mock.go -g
package grpcserver

import (
	"context"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"github.com/hydra13/shortify/api"
	"github.com/hydra13/shortify/internal/services/auth"
)

// Shorter - интерфейс для создания коротких URL
type Shorter interface {
	Create(ctx context.Context, long string) (string, error)
	CreateShortURL(shortID string) string
}

// UrlsKeeper - интерфейс для операций с URL
type UrlsKeeper interface {
	Get(ctx context.Context, shortURL string) (url string, found bool, err error)
	GetAllByUser(ctx context.Context, userID string) (urls map[string]string, err error)
}

// AuditService - интерфейс для аудита
type AuditService interface {
	PublishShortenEvent(longURL, userID string)
	PublishFollowEvent(longURL, userID string)
}

// Server - gRPC сервер
type Server struct {
	api.UnimplementedShortenerServiceServer
	shorter    Shorter
	urlsKeeper UrlsKeeper
	auth       *auth.AuthService
	audit      AuditService
	log        zerolog.Logger
}

// NewServer создает новый gRPC сервер
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

// Register регистрирует сервер на grpc.Server
func (s *Server) Register(grpcServer *grpc.Server) {
	api.RegisterShortenerServiceServer(grpcServer, s)
}
