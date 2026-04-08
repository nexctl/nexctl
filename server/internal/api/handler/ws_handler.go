package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nexctl/nexctl/server/internal/model"
	"github.com/nexctl/nexctl/server/internal/node"
	"github.com/nexctl/nexctl/server/internal/ws"
	"go.uber.org/zap"
)

// Agent WebSocket 凭据通过请求头传递，避免出现在 URL、访问日志与 Referer 中。
const (
	headerAgentID     = "X-NexCtl-Agent-Id"
	headerAgentSecret = "X-NexCtl-Agent-Secret"
)

func websocketOriginAllowed(r *http.Request, allowedOrigins []string) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	for _, o := range allowedOrigins {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}

// WSHandler handles agent websocket sessions.
type WSHandler struct {
	nodes    *node.Service
	ws       *ws.Service
	logger   *zap.Logger
	upgrader websocket.Upgrader
}

// NewWSHandler creates a websocket handler.
func NewWSHandler(nodes *node.Service, wsService *ws.Service, logger *zap.Logger, allowedOrigins []string) *WSHandler {
	allowed := append([]string(nil), allowedOrigins...)
	return &WSHandler{
		nodes:  nodes,
		ws:     wsService,
		logger: logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return websocketOriginAllowed(r, allowed)
			},
		},
	}
}

// AgentConnect upgrades an HTTP request into an agent websocket session.
func (h *WSHandler) AgentConnect(w http.ResponseWriter, r *http.Request) {
	agentID := strings.TrimSpace(r.Header.Get(headerAgentID))
	agentSecret := strings.TrimSpace(r.Header.Get(headerAgentSecret))
	if agentID == "" || agentSecret == "" {
		http.Error(w, "missing agent credentials", http.StatusUnauthorized)
		return
	}

	nodeRecord, appErr := h.nodes.AuthenticateAgent(r.Context(), agentID, agentSecret)
	if appErr != nil {
		http.Error(w, appErr.Message, http.StatusUnauthorized)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("upgrade websocket", zap.Error(err))
		return
	}

	conn.SetReadLimit(1 << 20)

	send, unregister := h.ws.AgentHub.Register(nodeRecord.ID)
	defer unregister()
	defer conn.Close()

	go ws.WritePump(conn, send)

	for {
		_ = conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		var message ws.Message
		if err := conn.ReadJSON(&message); err != nil {
			h.logger.Info("agent websocket disconnected", zap.Int64("node_id", nodeRecord.ID), zap.Error(err))
			return
		}

		switch message.Type {
		case ws.MessageTypeHeartbeat:
			h.handleAgentHeartbeat(r, send, nodeRecord, message)
		case ws.MessageTypeRuntimeState:
			h.handleAgentRuntimeState(r, send, nodeRecord, message)
		case ws.MessageTypeTerminalOutput:
			h.handleAgentTerminalOutput(message)
		case ws.MessageTypeTerminalExit:
			h.handleAgentTerminalExit(message)
		default:
			enqueueAgentSend(send, ws.Message{
				Type:      ws.MessageTypeError,
				RequestID: message.RequestID,
				Timestamp: time.Now().UTC(),
				Payload:   mustPayload(ws.ErrorPayload{MessageType: message.Type, Message: "unsupported message type"}),
			})
		}
	}
}

func enqueueAgentSend(send chan<- ws.Message, msg ws.Message) {
	select {
	case send <- msg:
	default:
	}
}

func (h *WSHandler) handleAgentTerminalOutput(message ws.Message) {
	var payload ws.TerminalOutputPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return
	}
	if strings.TrimSpace(payload.SessionID) == "" {
		return
	}
	_ = h.ws.TerminalBridge.DispatchFromAgent(payload.SessionID, message)
}

func (h *WSHandler) handleAgentTerminalExit(message ws.Message) {
	var payload ws.TerminalExitPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return
	}
	if strings.TrimSpace(payload.SessionID) == "" {
		return
	}
	fwd := ws.Message{
		Type:      ws.MessageTypeTerminalExit,
		Timestamp: time.Now().UTC(),
		Payload:   message.Payload,
	}
	_ = h.ws.TerminalBridge.DispatchFromAgent(payload.SessionID, fwd)
}

func (h *WSHandler) handleAgentHeartbeat(r *http.Request, send chan<- ws.Message, nodeRecord *model.Node, message ws.Message) {
	var payload ws.HeartbeatPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		enqueueAgentSend(send, ws.Message{
			Type:      ws.MessageTypeError,
			RequestID: message.RequestID,
			Timestamp: time.Now().UTC(),
			Payload:   mustPayload(ws.ErrorPayload{MessageType: ws.MessageTypeHeartbeat, Message: "invalid heartbeat payload"}),
		})
		return
	}
	if appErr := h.ws.HandleHeartbeat(r.Context(), nodeRecord, payload); appErr != nil {
		enqueueAgentSend(send, ws.Message{
			Type:      ws.MessageTypeError,
			RequestID: message.RequestID,
			Timestamp: time.Now().UTC(),
			Payload:   mustPayload(ws.ErrorPayload{MessageType: ws.MessageTypeHeartbeat, Message: appErr.Message}),
		})
		return
	}
	enqueueAgentSend(send, ws.Message{
		Type:      ws.MessageTypeAck,
		RequestID: message.RequestID,
		Timestamp: time.Now().UTC(),
		Payload:   mustPayload(ws.AckPayload{MessageType: ws.MessageTypeHeartbeat, Status: "ok"}),
	})
}

func (h *WSHandler) handleAgentRuntimeState(r *http.Request, send chan<- ws.Message, nodeRecord *model.Node, message ws.Message) {
	var payload ws.RuntimeStatePayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		enqueueAgentSend(send, ws.Message{
			Type:      ws.MessageTypeError,
			RequestID: message.RequestID,
			Timestamp: time.Now().UTC(),
			Payload:   mustPayload(ws.ErrorPayload{MessageType: ws.MessageTypeRuntimeState, Message: "invalid runtime_state payload"}),
		})
		return
	}
	if appErr := h.ws.HandleRuntimeState(r.Context(), nodeRecord, payload); appErr != nil {
		enqueueAgentSend(send, ws.Message{
			Type:      ws.MessageTypeError,
			RequestID: message.RequestID,
			Timestamp: time.Now().UTC(),
			Payload:   mustPayload(ws.ErrorPayload{MessageType: ws.MessageTypeRuntimeState, Message: appErr.Message}),
		})
		return
	}
	enqueueAgentSend(send, ws.Message{
		Type:      ws.MessageTypeAck,
		RequestID: message.RequestID,
		Timestamp: time.Now().UTC(),
		Payload:   mustPayload(ws.AckPayload{MessageType: ws.MessageTypeRuntimeState, Status: "ok"}),
	})
}

func mustPayload(v any) json.RawMessage {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return raw
}
