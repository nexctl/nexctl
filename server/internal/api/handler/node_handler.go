package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nexctl/nexctl/server/internal/api/middleware"
	"github.com/nexctl/nexctl/server/internal/node"
	"github.com/nexctl/nexctl/server/internal/runtime"
	"github.com/nexctl/nexctl/server/pkg/errcode"
	"github.com/nexctl/nexctl/server/pkg/response"
)

// NodeHandler handles node APIs.
type NodeHandler struct {
	nodes   *node.Service
	runtime *runtime.Service
}

// NewNodeHandler creates a node handler.
func NewNodeHandler(nodes *node.Service, runtime *runtime.Service) *NodeHandler {
	return &NodeHandler{nodes: nodes, runtime: runtime}
}

// Register handles node registration.
func (h *NodeHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req node.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, errcode.InvalidArgument, "invalid request body")
		return
	}

	resp, appErr := h.nodes.Register(r.Context(), req)
	if appErr != nil {
		status := http.StatusBadRequest
		if appErr.Code == errcode.Internal {
			status = http.StatusInternalServerError
		}
		response.WriteError(w, status, appErr.Code, appErr.Message)
		return
	}
	response.WriteCreated(w, resp)
}

// CreatePending pre-creates a node and returns an enrollment token for agent bootstrap.
func (h *NodeHandler) CreatePending(w http.ResponseWriter, r *http.Request) {
	var req node.CreatePendingNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, errcode.InvalidArgument, "invalid request body")
		return
	}

	claims, ok := middleware.UserClaimsFromContext(r.Context())
	actorID, actorName := "", ""
	if ok {
		actorID = strconv.FormatInt(claims.UserID, 10)
		actorName = claims.Username
	}

	resp, appErr := h.nodes.CreatePendingNode(r.Context(), req, actorID, actorName)
	if appErr != nil {
		status := http.StatusBadRequest
		if appErr.Code == errcode.Internal {
			status = http.StatusInternalServerError
		}
		response.WriteError(w, status, appErr.Code, appErr.Message)
		return
	}
	response.WriteCreated(w, resp)
}

// List handles node list.
func (h *NodeHandler) List(w http.ResponseWriter, r *http.Request) {
	resp, appErr := h.nodes.List(r.Context())
	if appErr != nil {
		response.WriteError(w, http.StatusInternalServerError, appErr.Code, appErr.Message)
		return
	}
	response.WriteOK(w, resp)
}

// Detail handles node detail.
func (h *NodeHandler) Detail(w http.ResponseWriter, r *http.Request) {
	nodeID, err := strconv.ParseInt(chi.URLParam(r, "nodeID"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errcode.InvalidArgument, "invalid node id")
		return
	}

	resp, appErr := h.nodes.GetDetail(r.Context(), nodeID)
	if appErr != nil {
		status := http.StatusInternalServerError
		if appErr.Code == errcode.NotFound {
			status = http.StatusNotFound
		}
		response.WriteError(w, status, appErr.Code, appErr.Message)
		return
	}
	response.WriteOK(w, resp)
}

// RegenerateEnrollmentToken issues a new enrollment token for a pending node so the console can show install commands.
func (h *NodeHandler) RegenerateEnrollmentToken(w http.ResponseWriter, r *http.Request) {
	nodeID, err := strconv.ParseInt(chi.URLParam(r, "nodeID"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errcode.InvalidArgument, "invalid node id")
		return
	}

	claims, ok := middleware.UserClaimsFromContext(r.Context())
	actorID, actorName := "", ""
	if ok {
		actorID = strconv.FormatInt(claims.UserID, 10)
		actorName = claims.Username
	}

	resp, appErr := h.nodes.RegenerateEnrollmentToken(r.Context(), nodeID, actorID, actorName)
	if appErr != nil {
		status := http.StatusInternalServerError
		switch appErr.Code {
		case errcode.NotFound:
			status = http.StatusNotFound
		case errcode.InvalidArgument:
			status = http.StatusBadRequest
		}
		response.WriteError(w, status, appErr.Code, appErr.Message)
		return
	}
	response.WriteOK(w, resp)
}

// Delete removes a node by ID.
func (h *NodeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	nodeID, err := strconv.ParseInt(chi.URLParam(r, "nodeID"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errcode.InvalidArgument, "invalid node id")
		return
	}

	claims, ok := middleware.UserClaimsFromContext(r.Context())
	actorID, actorName := "", ""
	if ok {
		actorID = strconv.FormatInt(claims.UserID, 10)
		actorName = claims.Username
	}

	if appErr := h.nodes.Delete(r.Context(), nodeID, actorID, actorName); appErr != nil {
		status := http.StatusInternalServerError
		if appErr.Code == errcode.NotFound {
			status = http.StatusNotFound
		}
		response.WriteError(w, status, appErr.Code, appErr.Message)
		return
	}
	response.WriteOK(w, map[string]any{"deleted": true})
}

// UpdateRuntimeState handles node runtime state update.
func (h *NodeHandler) UpdateRuntimeState(w http.ResponseWriter, r *http.Request) {
	nodeID, err := strconv.ParseInt(chi.URLParam(r, "nodeID"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errcode.InvalidArgument, "invalid node id")
		return
	}

	var req runtime.UpdateStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, errcode.InvalidArgument, "invalid request body")
		return
	}

	if appErr := h.runtime.Update(r.Context(), nodeID, req); appErr != nil {
		response.WriteError(w, http.StatusInternalServerError, appErr.Code, appErr.Message)
		return
	}
	response.WriteOK(w, map[string]any{"updated": true})
}

// CurrentUser returns the authenticated operator context.
func (h *NodeHandler) CurrentUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserClaimsFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, errcode.Unauthorized, "unauthorized")
		return
	}
	response.WriteOK(w, claims)
}
