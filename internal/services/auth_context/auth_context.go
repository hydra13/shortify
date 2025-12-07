package authcontext

import "context"

type ContextKey string

const (
	ContextUserIDKey    ContextKey = "userID"
	ContextIsNewUserKey ContextKey = "isNew"
)

func CreateContextWithUserID(ctx context.Context, userID string, isNew bool) context.Context {
	ctx = context.WithValue(ctx, ContextUserIDKey, userID)
	return context.WithValue(ctx, ContextIsNewUserKey, isNew)
}

func GetUserIDFromContext(ctx context.Context) (userID string, isNew bool) {
	userIDInterface := ctx.Value(ContextUserIDKey)
	isNewInterface := ctx.Value(ContextIsNewUserKey)

	if userIDInterface != nil {
		userID = userIDInterface.(string)
	}

	if isNewInterface != nil {
		isNew = isNewInterface.(bool)
	}

	return
}
