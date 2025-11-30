package auth

import (
	"net/http"

	"github.com/rs/zerolog"

	authContext "github.com/hydra13/shortify/internal/services/auth_context"
)

type AuthService interface {
	GetOrCreateUser(request *http.Request) (userID string, inNew bool)
	SetAuthCookie(w http.ResponseWriter, userID string)
}

func NewAuthMiddleware(as AuthService, log zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, isNew := as.GetOrCreateUser(r)

			log.Debug().
				Str("user_id", userID).
				Bool("is_new", isNew).
				Msg("auth middleware")

			if isNew {
				as.SetAuthCookie(w, userID)
			}

			ctx := authContext.CreateContextWithUserID(r.Context(), userID, isNew)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
