package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/nexctl/nexctl/server/pkg/errcode"
	"github.com/nexctl/nexctl/server/pkg/response"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles health endpoints.
type HealthHandler struct {
	db  *sql.DB
	rdb *redis.Client
}

// NewHealthHandler creates a health handler.
func NewHealthHandler(db *sql.DB, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb}
}

// Healthz reports dependency health.
func (h *HealthHandler) Healthz(w http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		response.WriteError(w, http.StatusServiceUnavailable, errcode.Internal, "mysql unavailable")
		return
	}
	if err := h.rdb.Ping(ctx).Err(); err != nil {
		response.WriteError(w, http.StatusServiceUnavailable, errcode.Internal, "redis unavailable")
		return
	}
	response.WriteOK(w, map[string]string{"status": "ok"})
}
