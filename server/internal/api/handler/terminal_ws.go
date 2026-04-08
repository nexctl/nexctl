package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nexctl/nexctl/server/internal/api/middleware"
	"github.com/nexctl/nexctl/server/internal/config"
	"github.com/nexctl/nexctl/server/internal/ws"
	"github.com/nexctl/nexctl/server/pkg/jwtutil"
	"go.uber.org/zap"
)

// TerminalWSHandler 浏览器与节点终端之间的 WebSocket（经控制面转发至 Agent）。
type TerminalWSHandler struct {
	cfg   config.AuthConfig
	wsSvc *ws.Service
	logger *zap.Logger
	up    websocket.Upgrader
}

// NewTerminalWSHandler 创建终端 WebSocket 处理器。
func NewTerminalWSHandler(cfg config.AuthConfig, wsSvc *ws.Service, logger *zap.Logger, allowedOrigins []string) *TerminalWSHandler {
	allowed := append([]string(nil), allowedOrigins...)
	return &TerminalWSHandler{
		cfg:   cfg,
		wsSvc: wsSvc,
		logger: logger,
		up: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return websocketOriginAllowed(r, allowed)
			},
		},
	}
}

// ServeWS 需经 JWT 鉴权（query: token=）与 nodes:write 权限；将会话与 Agent 上的 PTY 绑定。
func (h *TerminalWSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	claims, err := jwtutil.Parse(h.cfg.JWTSecret, token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	if !middleware.RoleAllowsPermission(claims.RoleCode, "nodes:write") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	nodeIDStr := strings.TrimSpace(chi.URLParam(r, "nodeID"))
	nodeID, err := strconv.ParseInt(nodeIDStr, 10, 64)
	if err != nil || nodeID <= 0 {
		http.Error(w, "invalid node id", http.StatusBadRequest)
		return
	}

	if !h.wsSvc.AgentHub.Online(nodeID) {
		http.Error(w, "agent offline", http.StatusServiceUnavailable)
		return
	}

	conn, err := h.up.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("terminal ws upgrade", zap.Error(err))
		return
	}
	defer conn.Close()

	sessionID := uuid.NewString()
	cols, rows := 120, 40
	if c := r.URL.Query().Get("cols"); c != "" {
		if v, e := strconv.Atoi(c); e == nil && v > 0 && v < 512 {
			cols = v
		}
	}
	if rws := r.URL.Query().Get("rows"); rws != "" {
		if v, e := strconv.Atoi(rws); e == nil && v > 0 && v < 512 {
			rows = v
		}
	}

	// 必须先注册桥接并启动写泵，再向 Agent 下发 terminal_open，否则 Agent 首包 terminal_output 可能在 Register 之前到达并被丢弃。
	toBrowser, unregisterBrowser := h.wsSvc.TerminalBridge.Register(sessionID, nodeID)
	defer unregisterBrowser()

	go ws.BrowserWritePump(conn, toBrowser)

	openPayload, _ := json.Marshal(ws.TerminalOpenPayload{
		SessionID: sessionID,
		Cols:      cols,
		Rows:      rows,
	})
	openMsg := ws.Message{
		Type:      ws.MessageTypeTerminalOpen,
		RequestID: sessionID,
		Timestamp: time.Now().UTC(),
		Payload:   openPayload,
	}
	if err := h.wsSvc.AgentHub.Send(nodeID, openMsg); err != nil {
		_ = conn.WriteJSON(ws.Message{
			Type:      ws.MessageTypeTerminalError,
			Timestamp: time.Now().UTC(),
			Payload:   mustPayload(ws.TerminalErrorPayload{Message: err.Error()}),
		})
		return
	}

	readDeadline := time.Now().Add(300 * time.Second)
	_ = conn.SetReadDeadline(readDeadline)

	for {
		var fromBrowser ws.Message
		if err := conn.ReadJSON(&fromBrowser); err != nil {
			_ = h.closeTerminalSession(nodeID, sessionID)
			return
		}
		_ = conn.SetReadDeadline(time.Now().Add(300 * time.Second))

		switch fromBrowser.Type {
		case ws.MessageTypeTerminalInput:
			var p ws.TerminalInputPayload
			if err := json.Unmarshal(fromBrowser.Payload, &p); err != nil {
				continue
			}
			p.SessionID = sessionID
			b, _ := json.Marshal(p)
			_ = h.wsSvc.AgentHub.Send(nodeID, ws.Message{
				Type:      ws.MessageTypeTerminalInput,
				RequestID: fromBrowser.RequestID,
				Timestamp: time.Now().UTC(),
				Payload:   b,
			})
		case ws.MessageTypeTerminalResize:
			var p ws.TerminalResizePayload
			if err := json.Unmarshal(fromBrowser.Payload, &p); err != nil {
				continue
			}
			p.SessionID = sessionID
			b, _ := json.Marshal(p)
			_ = h.wsSvc.AgentHub.Send(nodeID, ws.Message{
				Type:      ws.MessageTypeTerminalResize,
				RequestID: fromBrowser.RequestID,
				Timestamp: time.Now().UTC(),
				Payload:   b,
			})
		case ws.MessageTypeTerminalClose:
			_ = h.closeTerminalSession(nodeID, sessionID)
			return
		default:
		}
	}
}

func (h *TerminalWSHandler) closeTerminalSession(nodeID int64, sessionID string) error {
	p, _ := json.Marshal(ws.TerminalClosePayload{SessionID: sessionID})
	return h.wsSvc.AgentHub.Send(nodeID, ws.Message{
		Type:      ws.MessageTypeTerminalClose,
		Timestamp: time.Now().UTC(),
		Payload:   p,
	})
}
