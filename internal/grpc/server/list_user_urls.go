package grpcserver

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/hydra13/shortify/api"
)

func (s *Server) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*api.UserURLsResponse, error) {
	userID, err := s.getUserIDFromMetadata(ctx)
	if err != nil {
		s.log.Debug().Err(err).Msg("ListUserURLs: error getting user from metadata")
		return nil, err
	}

	urls, err := s.urlsKeeper.GetAllByUser(ctx, userID)
	if err != nil {
		s.log.Error().Str("user_id", userID).Err(err).Msg("ListUserURLs: error getting urls")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	urlData := make([]*api.URLData, 0, len(urls))
	for shortID, originalURL := range urls {
		urlData = append(urlData, &api.URLData{
			ShortUrl:    s.shorter.CreateShortURL(shortID),
			OriginalUrl: originalURL,
		})
	}

	return &api.UserURLsResponse{Url: urlData}, nil
}
