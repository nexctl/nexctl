package ws

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/nexctl/nexctl/server/internal/audit"
	"github.com/nexctl/nexctl/server/internal/config"
	"github.com/nexctl/nexctl/server/internal/model"
	"github.com/nexctl/nexctl/server/internal/repository"
	"github.com/nexctl/nexctl/server/internal/runtime"
	"github.com/nexctl/nexctl/server/pkg/errcode"
	"go.uber.org/zap"
)

const (
	// MessageTypeHeartbeat is the heartbeat websocket message type.
	MessageTypeHeartbeat = "heartbeat"
	// MessageTypeRuntimeState is the runtime-state websocket message type.
	MessageTypeRuntimeState = "runtime_state"
	// MessageTypeAck is the generic acknowledgement message type.
	MessageTypeAck = "ack"
	// MessageTypeError is the generic error message type.
	MessageTypeError = "error"
	// MessageTypeTaskDispatch is reserved for future task dispatch.
	MessageTypeTaskDispatch = "task_dispatch"
	// MessageTypeFileDispatch is reserved for future file operations.
	MessageTypeFileDispatch = "file_dispatch"
	MessageTypeUpgradeCommand = "upgrade_command"

	// 终端（浏览器 <-> 控制面 <-> Agent PTY）
	MessageTypeTerminalOpen    = "terminal_open"
	MessageTypeTerminalInput   = "terminal_input"
	MessageTypeTerminalResize  = "terminal_resize"
	MessageTypeTerminalClose   = "terminal_close"
	MessageTypeTerminalOutput  = "terminal_output"
	MessageTypeTerminalExit    = "terminal_exit"
	MessageTypeTerminalError   = "terminal_error"
)

// TerminalOpenPayload 控制面 -> Agent：创建 PTY 会话。
type TerminalOpenPayload struct {
	SessionID string `json:"session_id"`
	Cols      int    `json:"cols"`
	Rows      int    `json:"rows"`
}

// TerminalInputPayload 控制面 -> Agent：用户键盘输入。
type TerminalInputPayload struct {
	SessionID string `json:"session_id"`
	Data      string `json:"data"` // base64
}

// TerminalResizePayload 控制面 -> Agent：窗口大小。
type TerminalResizePayload struct {
	SessionID string `json:"session_id"`
	Cols      int    `json:"cols"`
	Rows      int    `json:"rows"`
}

// TerminalClosePayload 控制面 -> Agent 或任一端：关闭会话。
type TerminalClosePayload struct {
	SessionID string `json:"session_id"`
}

// TerminalOutputPayload Agent -> 控制面：PTY 输出。
type TerminalOutputPayload struct {
	SessionID string `json:"session_id"`
	Data      string `json:"data"` // base64
}

// TerminalExitPayload Agent -> 控制面：子进程退出。
type TerminalExitPayload struct {
	SessionID string `json:"session_id"`
	Code      int    `json:"code"`
}

// TerminalErrorPayload 控制面 -> 浏览器：错误说明。
type TerminalErrorPayload struct {
	Message string `json:"message"`
}

// UpgradeCommandPayload 控制面 -> Agent：触发一次 GitHub Release 自更新检查。
type UpgradeCommandPayload struct {
	Source string `json:"source,omitempty"` // 例如 console
}

// Message is the base websocket envelope.
type Message struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// HeartbeatPayload is the heartbeat websocket payload.
type HeartbeatPayload struct {
	SentAt time.Time `json:"sent_at"`
}

// RuntimeStatePayload is the runtime-state websocket payload.
type RuntimeStatePayload struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskPercent   float64 `json:"disk_percent"`
	NetworkRxBps  uint64  `json:"network_rx_bps"`
	NetworkTxBps  uint64  `json:"network_tx_bps"`
	Load1         float64 `json:"load_1"`
	Load5         float64 `json:"load_5"`
	Load15        float64 `json:"load_15"`
	UptimeSeconds uint64  `json:"uptime_seconds"`
	ProcessCount  uint32  `json:"process_count"`
	Timestamp     string  `json:"timestamp,omitempty"`
}

// AckPayload is the generic websocket acknowledgement payload.
type AckPayload struct {
	MessageType string `json:"message_type"`
	Status      string `json:"status"`
}

// ErrorPayload is the generic websocket error payload.
type ErrorPayload struct {
	MessageType string `json:"message_type"`
	Message     string `json:"message"`
}

// Service implements websocket message handling.
type Service struct {
	cfg     config.NodeConfig
	nodes   repository.NodeRepository
	runtime *runtime.Service
	cache   repository.NodeSessionCache
	audit   *audit.Service
	logger  *zap.Logger

	AgentHub       *AgentHub
	TerminalBridge *TerminalBridge
}

// NewService creates a websocket service.
func NewService(cfg config.NodeConfig, nodes repository.NodeRepository, runtime *runtime.Service, cache repository.NodeSessionCache, audit *audit.Service, logger *zap.Logger) *Service {
	return &Service{
		cfg:            cfg,
		nodes:          nodes,
		runtime:        runtime,
		cache:          cache,
		audit:          audit,
		logger:         logger,
		AgentHub:       NewAgentHub(),
		TerminalBridge: NewTerminalBridge(),
	}
}

// HandleHeartbeat handles an agent heartbeat message.
func (s *Service) HandleHeartbeat(ctx context.Context, node *model.Node, payload HeartbeatPayload) *errcode.AppError {
	now := time.Now().UTC()
	if !payload.SentAt.IsZero() {
		now = payload.SentAt.UTC()
	}
	if err := s.nodes.UpdateHeartbeat(ctx, node.ID, now, model.NodeStatusOnline); err != nil {
		return errcode.Wrap(errcode.Internal, "update heartbeat failed", err)
	}
	if err := s.cache.MarkOnline(ctx, node.ID, time.Duration(s.cfg.HeartbeatTimeoutSeconds)*time.Second); err != nil {
		s.logger.Warn("cache node online state", zap.Int64("node_id", node.ID), zap.Error(err))
	}
	return nil
}

// HandleRuntimeState handles an agent runtime state message.
func (s *Service) HandleRuntimeState(ctx context.Context, node *model.Node, payload RuntimeStatePayload) *errcode.AppError {
	appErr := s.runtime.Update(ctx, node.ID, runtime.UpdateStateRequest{
		CPUPercent:    payload.CPUPercent,
		MemoryPercent: payload.MemoryPercent,
		DiskPercent:   payload.DiskPercent,
		NetworkRxBps:  payload.NetworkRxBps,
		NetworkTxBps:  payload.NetworkTxBps,
		Load1:         payload.Load1,
		Load5:         payload.Load5,
		Load15:        payload.Load15,
		UptimeSeconds: payload.UptimeSeconds,
		ProcessCount:  payload.ProcessCount,
		Timestamp:     payload.Timestamp,
	})
	if appErr != nil {
		return appErr
	}

	_ = s.audit.Record(ctx, audit.RecordInput{
		ActorType:    "agent",
		ActorID:      node.AgentID,
		ActorName:    node.Name,
		Action:       "node.runtime_state.update",
		ResourceType: "node",
		ResourceID:   strconv.FormatInt(node.ID, 10),
		Detail:       `{"source":"ws"}`,
	})
	return nil
}
