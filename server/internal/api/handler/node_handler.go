package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nexctl/nexctl/server/internal/api/middleware"
	"github.com/nexctl/nexctl/server/internal/node"
	"github.com/nexctl/nexctl/server/internal/runtime"
	"github.com/nexctl/nexctl/server/internal/ws"
	"github.com/nexctl/nexctl/server/pkg/errcode"
	"github.com/nexctl/nexctl/server/pkg/response"
)

// NodeHandler handles node APIs.
type NodeHandler struct {
	nodes   *node.Service
	runtime *runtime.Service
	wsSvc   *ws.Service
}

// NewNodeHandler creates a node handler.
func NewNodeHandler(nodes *node.Service, runtime *runtime.Service, wsSvc *ws.Service) *NodeHandler {
	return &NodeHandler{nodes: nodes, runtime: runtime, wsSvc: wsSvc}
}

// CreatePending 创建节点并返回固定的 agent_id / agent_secret / node_key。
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

// AgentCredentials 返回节点的固定接入凭据（用于控制台展示安装命令）。
func (h *NodeHandler) AgentCredentials(w http.ResponseWriter, r *http.Request) {
	nodeID, err := strconv.ParseInt(chi.URLParam(r, "nodeID"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errcode.InvalidArgument, "invalid node id")
		return
	}

	resp, appErr := h.nodes.GetNodeAgentCredentials(r.Context(), nodeID)
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

// TriggerAgentUpgrade 向在线 Agent 下发 upgrade_command，由 Agent 执行一次 GitHub 自更新检查。
func (h *NodeHandler) TriggerAgentUpgrade(w http.ResponseWriter, r *http.Request) {
	nodeID, err := strconv.ParseInt(chi.URLParam(r, "nodeID"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errcode.InvalidArgument, "invalid node id")
		return
	}

	if _, appErr := h.nodes.GetDetail(r.Context(), nodeID); appErr != nil {
		status := http.StatusInternalServerError
		if appErr.Code == errcode.NotFound {
			status = http.StatusNotFound
		}
		response.WriteError(w, status, appErr.Code, appErr.Message)
		return
	}

	if !h.wsSvc.AgentHub.Online(nodeID) {
		response.WriteError(w, http.StatusServiceUnavailable, errcode.Internal, "agent offline")
		return
	}

	payload, _ := json.Marshal(ws.UpgradeCommandPayload{Source: "console"})
	msg := ws.Message{
		Type:      ws.MessageTypeUpgradeCommand,
		RequestID: uuid.NewString(),
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}
	if err := h.wsSvc.AgentHub.Send(nodeID, msg); err != nil {
		response.WriteError(w, http.StatusServiceUnavailable, errcode.Internal, err.Error())
		return
	}
	response.WriteOK(w, map[string]any{"queued": true, "request_id": msg.RequestID})
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
