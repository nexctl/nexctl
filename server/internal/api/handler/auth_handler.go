package handler

import (
	"encoding/json"
	"net/http"

	"github.com/nexctl/nexctl/server/internal/auth"
	"github.com/nexctl/nexctl/server/pkg/errcode"
	"github.com/nexctl/nexctl/server/pkg/response"
)

// AuthHandler handles authentication APIs.
type AuthHandler struct {
	service *auth.Service
}

// NewAuthHandler creates an auth handler.
func NewAuthHandler(service *auth.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

// Login handles user login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, errcode.InvalidArgument, "invalid request body")
		return
	}

	resp, appErr := h.service.Login(r.Context(), req)
	if appErr != nil {
		status := http.StatusUnauthorized
		if appErr.Code == errcode.Internal {
			status = http.StatusInternalServerError
		}
		response.WriteError(w, status, appErr.Code, appErr.Message)
		return
	}
	response.WriteOK(w, resp)
}
