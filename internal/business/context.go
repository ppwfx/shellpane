package business

import "context"

type userIDKey struct{}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func UserID(ctx context.Context) string {
	v := ctx.Value(userIDKey{})
	if v == nil {
		return ""
	}

	return v.(string)
}
