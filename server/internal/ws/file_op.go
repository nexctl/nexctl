package ws

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// FileDispatchPayload 控制面 -> Agent：远程文件操作（与 Agent filemgr.DispatchPayload 字段一致）。
type FileDispatchPayload struct {
	Op         string `json:"op"`
	Path       string `json:"path"`
	PathTo     string `json:"path_to,omitempty"`
	ContentB64 string `json:"content_b64,omitempty"`
	MaxBytes   int    `json:"max_bytes,omitempty"`
	Recursive  bool   `json:"recursive,omitempty"`
}

// FileEntry 目录列表单项。
type FileEntry struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

// FileReportPayload Agent -> 控制面：file_dispatch 执行结果。
type FileReportPayload struct {
	OK         bool        `json:"ok"`
	Error      string      `json:"error,omitempty"`
	Entries    []FileEntry `json:"entries,omitempty"`
	ContentB64 string      `json:"content_b64,omitempty"`
	Size       int64       `json:"size,omitempty"`
	ModTime    string      `json:"mod_time,omitempty"`
	IsDir      bool        `json:"is_dir,omitempty"`
}

// FileOpRegistry 同步等待 Agent 返回的 file_report（按 request_id 关联）。
type FileOpRegistry struct {
	mu sync.Mutex
	m  map[string]chan FileReportPayload
}

// NewFileOpRegistry creates a registry.
func NewFileOpRegistry() *FileOpRegistry {
	return &FileOpRegistry{m: make(map[string]chan FileReportPayload)}
}

func (r *FileOpRegistry) register(id string) chan FileReportPayload {
	ch := make(chan FileReportPayload, 1)
	r.mu.Lock()
	r.m[id] = ch
	r.mu.Unlock()
	return ch
}

func (r *FileOpRegistry) unregister(id string) {
	r.mu.Lock()
	delete(r.m, id)
	r.mu.Unlock()
}

// Complete 由 WS 读循环在收到 file_report 时调用。
func (r *FileOpRegistry) Complete(id string, p FileReportPayload) {
	r.mu.Lock()
	ch, ok := r.m[id]
	r.mu.Unlock()
	if !ok {
		return
	}
	select {
	case ch <- p:
	default:
	}
}

// ExecuteFileOp 向在线 Agent 发送 file_dispatch 并同步等待 file_report。
func (s *Service) ExecuteFileOp(ctx context.Context, nodeID int64, payload FileDispatchPayload) (FileReportPayload, error) {
	if !s.AgentHub.Online(nodeID) {
		return FileReportPayload{}, ErrAgentOffline
	}
	reqID := uuid.NewString()
	ch := s.fileOps.register(reqID)
	defer s.fileOps.unregister(reqID)

	raw, err := json.Marshal(payload)
	if err != nil {
		return FileReportPayload{}, err
	}
	msg := Message{
		Type:      MessageTypeFileDispatch,
		RequestID: reqID,
		Timestamp: time.Now().UTC(),
		Payload:   raw,
	}
	if err := s.AgentHub.Send(nodeID, msg); err != nil {
		return FileReportPayload{}, err
	}
	select {
	case <-ctx.Done():
		return FileReportPayload{}, ctx.Err()
	case res := <-ch:
		return res, nil
	case <-time.After(90 * time.Second):
		return FileReportPayload{}, errors.New("file operation timeout")
	}
}

// CompleteFileOp WS 收到 Agent 的 file_report 时调用。
func (s *Service) CompleteFileOp(requestID string, p FileReportPayload) {
	if requestID == "" {
		return
	}
	s.fileOps.Complete(requestID, p)
}
