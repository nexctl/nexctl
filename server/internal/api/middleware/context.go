package middleware

import "context"

type contextKey string

const (
	userClaimsKey contextKey = "user_claims"
)

// UserClaims is the authenticated user identity inside request context.
type UserClaims struct {
	UserID   int64
	Username string
	RoleCode string
}

// WithUserClaims stores authenticated user claims in context.
func WithUserClaims(ctx context.Context, claims UserClaims) context.Context {
	return context.WithValue(ctx, userClaimsKey, claims)
}

// UserClaimsFromContext extracts authenticated user claims from context.
func UserClaimsFromContext(ctx context.Context) (UserClaims, bool) {
	value := ctx.Value(userClaimsKey)
	claims, ok := value.(UserClaims)
	return claims, ok
}
