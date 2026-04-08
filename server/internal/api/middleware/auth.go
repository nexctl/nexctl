package middleware

import (
	"net/http"
	"strings"

	"github.com/nexctl/nexctl/server/internal/config"
	"github.com/nexctl/nexctl/server/pkg/errcode"
	"github.com/nexctl/nexctl/server/pkg/jwtutil"
	"github.com/nexctl/nexctl/server/pkg/response"
)

func roleAllowsPermission(roleCode, required string) bool {
	if required == "" {
		return true
	}
	r := strings.ToLower(strings.TrimSpace(roleCode))
	switch r {
	case "admin", "super_admin", "root":
		return true
	case "viewer", "readonly":
		switch required {
		case "nodes:read", "modules:read":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

// AuthMiddleware validates bearer JWT and attaches user claims.
type AuthMiddleware struct {
	cfg config.AuthConfig
}

// NewAuthMiddleware creates an auth middleware.
func NewAuthMiddleware(cfg config.AuthConfig) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg}
}

// RequireLogin enforces JWT authentication.
func (m *AuthMiddleware) RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.WriteError(w, http.StatusUnauthorized, errcode.Unauthorized, "missing bearer token")
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtutil.Parse(m.cfg.JWTSecret, tokenString)
		if err != nil {
			response.WriteError(w, http.StatusUnauthorized, errcode.Unauthorized, "invalid token")
			return
		}

		ctx := WithUserClaims(r.Context(), UserClaims{
			UserID:   claims.UserID,
			Username: claims.Username,
			RoleCode: claims.RoleCode,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission enforces RBAC for the given permission code (e.g. nodes:read).
func (m *AuthMiddleware) RequirePermission(required string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := UserClaimsFromContext(r.Context())
			if !ok {
				response.WriteError(w, http.StatusUnauthorized, errcode.Unauthorized, "unauthorized")
				return
			}
			if !roleAllowsPermission(claims.RoleCode, required) {
				response.WriteError(w, http.StatusForbidden, errcode.Forbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
