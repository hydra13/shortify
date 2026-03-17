package grpcserver

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hydra13/shortify/api"
	"github.com/hydra13/shortify/internal/models"
	authContext "github.com/hydra13/shortify/internal/services/auth_context"
)

func (s *Server) ShortenURL(ctx context.Context, req *api.URLShortenRequest) (*api.URLShortenResponse, error) {
	if req.Url == "" {
		s.log.Debug().Msg("ShortenURL: empty url in request")
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	userID, err := s.getUserIDFromMetadata(ctx)
	if err != nil {
		s.log.Debug().Err(err).Msg("ShortenURL: error getting user from metadata")
	}

	ctx = authContext.CreateContextWithUserID(ctx, userID, false)

	shortURL, err := s.shorter.Create(ctx, req.Url)
	if err != nil {
		if errors.Is(err, models.ErrValidation) {
			s.log.Debug().Str("input_url", req.Url).Msg("ShortenURL: validation error")
			return nil, status.Error(codes.InvalidArgument, "invalid url")
		}

		if errors.Is(err, models.ErrConflict) {
			s.log.Debug().Str("input_url", req.Url).Msg("ShortenURL: url already exists")

			return &api.URLShortenResponse{Result: shortURL}, nil
		}

		if errors.Is(err, models.ErrInternal) {
			s.log.Error().Str("input_url", req.Url).Err(err).Msg("ShortenURL: internal error")
			return nil, status.Error(codes.Internal, "internal server error")
		}

		s.log.Error().Str("input_url", req.Url).Err(err).Msg("ShortenURL: unhandled error")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	s.audit.PublishShortenEvent(req.Url, userID)

	return &api.URLShortenResponse{Result: shortURL}, nil
}
