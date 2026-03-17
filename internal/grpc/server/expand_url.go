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

func (s *Server) ExpandURL(ctx context.Context, req *api.URLExpandRequest) (*api.URLExpandResponse, error) {
	if req.Id == "" {
		s.log.Debug().Msg("ExpandURL: empty id in request")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

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

	s.audit.PublishFollowEvent(url, userID)

	return &api.URLExpandResponse{Result: url}, nil
}
