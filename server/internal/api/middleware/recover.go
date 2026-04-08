package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/nexctl/nexctl/server/pkg/errcode"
	"github.com/nexctl/nexctl/server/pkg/response"
	"go.uber.org/zap"
)

// Recover converts panics into structured 500 responses.
func Recover(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("panic recovered",
						zap.Any("panic", recovered),
						zap.ByteString("stack", debug.Stack()),
					)
					response.WriteError(w, http.StatusInternalServerError, errcode.Internal, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
