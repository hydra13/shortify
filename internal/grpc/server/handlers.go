package grpcserver

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/hydra13/shortify/api"
	"github.com/hydra13/shortify/internal/models"
	authContext "github.com/hydra13/shortify/internal/services/auth_context"
)

// Ensure zerolog is explicitly used
var _ = zerolog.Logger{}

// ShortenURL создает короткий URL
func (s *Server) ShortenURL(ctx context.Context, req *api.URLShortenRequest) (*api.URLShortenResponse, error) {
	if req.Url == "" {
		s.log.Debug().Msg("ShortenURL: empty url in request")
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	// Извлекаем userID из metadata
	userID, err := s.getUserIDFromMetadata(ctx)
	if err != nil {
		s.log.Debug().Err(err).Msg("ShortenURL: error getting user from metadata")
		// Не прерываем выполнение, продолжаем с пустым userID
	}

	// Создаем контекст с userID
	ctx = authContext.CreateContextWithUserID(ctx, userID, false)

	shortURL, err := s.shorter.Create(ctx, req.Url)
	if err != nil {
		if errors.Is(err, models.ErrValidation) {
			s.log.Debug().Str("input_url", req.Url).Msg("ShortenURL: validation error")
			return nil, status.Error(codes.InvalidArgument, "invalid url")
		}

		if errors.Is(err, models.ErrConflict) {
			s.log.Debug().Str("input_url", req.Url).Msg("ShortenURL: url already exists")
			// Возвращаем уже существующий короткий URL
			return &api.URLShortenResponse{Result: shortURL}, nil
		}

		if errors.Is(err, models.ErrInternal) {
			s.log.Error().Str("input_url", req.Url).Err(err).Msg("ShortenURL: internal error")
			return nil, status.Error(codes.Internal, "internal server error")
		}

		s.log.Error().Str("input_url", req.Url).Err(err).Msg("ShortenURL: unhandled error")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	// Публикуем событие аудита
	s.audit.PublishShortenEvent(req.Url, userID)

	return &api.URLShortenResponse{Result: shortURL}, nil
}

// ExpandURL раскрывает короткий URL
func (s *Server) ExpandURL(ctx context.Context, req *api.URLExpandRequest) (*api.URLExpandResponse, error) {
	if req.Id == "" {
		s.log.Debug().Msg("ExpandURL: empty id in request")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Извлекаем userID из metadata для аудита
	userID, _ := s.getUserIDFromMetadata(ctx)
	ctx = authContext.CreateContextWithUserID(ctx, userID, false)

	url, found, err := s.urlsKeeper.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, models.ErrURLIsDeleted) {
			s.log.Debug().Str("short_url_key", req.Id).Msg("ExpandURL: url is deleted")
			return nil, status.Error(codes.NotFound, "url is deleted")
		}

		s.log.Error().Str("short_url_key", req.Id).Err(err).Msg("ExpandURL: error getting url")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	if !found {
		s.log.Debug().Str("short_url_key", req.Id).Msg("ExpandURL: url not found")
		return nil, status.Error(codes.NotFound, "url not found")
	}

	// Публикуем событие следования за URL
	s.audit.PublishFollowEvent(url, userID)

	return &api.URLExpandResponse{Result: url}, nil
}

// ListUserURLs возвращает все URL пользователя
func (s *Server) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*api.UserURLsResponse, error) {
	// Извлекаем userID из metadata
	userID, err := s.getUserIDFromMetadata(ctx)
	if err != nil {
		s.log.Debug().Err(err).Msg("ListUserURLs: error getting user from metadata")
		return nil, err // Return the original error with proper gRPC status
	}

	// Получаем URL пользователя
	urls, err := s.urlsKeeper.GetAllByUser(ctx, userID)
	if err != nil {
		s.log.Error().Str("user_id", userID).Err(err).Msg("ListUserURLs: error getting urls")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	// Преобразуем в формат ответа
	urlData := make([]*api.URLData, 0, len(urls))
	for shortID, originalURL := range urls {
		urlData = append(urlData, &api.URLData{
			ShortUrl:    s.shorter.CreateShortURL(shortID), // Полный URL
			OriginalUrl: originalURL,
		})
	}

	return &api.UserURLsResponse{Url: urlData}, nil
}

// getUserIDFromMetadata извлекает userID из gRPC metadata
func (s *Server) getUserIDFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata not found")
	}

	authValues := md.Get("authorization")
	if len(authValues) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization header not found")
	}

	// Парсим токен (формат: "Bearer <token>" или просто "<token>")
	token := authValues[0]
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Валидируем токен
	userID, err := s.auth.GetUserIDFromToken(token)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid token")
	}

	return userID, nil
}
