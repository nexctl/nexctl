package ws

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

// ErrAgentOffline 表示该节点当前无已连接的 Agent WebSocket。
var ErrAgentOffline = errors.New("agent offline")

const agentSendQueue = 256

// AgentHub 登记每个节点上行的 Agent WebSocket，并向其下发控制消息（如终端）。
type AgentHub struct {
	mu     sync.RWMutex
	agents map[int64]*agentEntry
}

type agentEntry struct {
	send chan Message
	conn *websocket.Conn
}

func NewAgentHub() *AgentHub {
	return &AgentHub{agents: make(map[int64]*agentEntry)}
}

// Register 注册节点 Agent；同一 nodeID 新连接会替换旧连接。
// 会先关闭旧 WebSocket，再关闭旧下发 chan，避免旧读循环仍向已关闭的 chan 发送而 panic。
func (h *AgentHub) Register(nodeID int64, conn *websocket.Conn) (send chan Message, unregister func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := h.agents[nodeID]; ok && old != nil {
		if old.conn != nil {
			_ = old.conn.Close()
		}
		close(old.send)
	}
	ch := make(chan Message, agentSendQueue)
	h.agents[nodeID] = &agentEntry{send: ch, conn: conn}
	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if cur, ok := h.agents[nodeID]; ok && cur != nil && cur.send == ch {
			delete(h.agents, nodeID)
			close(ch)
		}
	}
}

// Send 向节点上的 Agent 下发一条 JSON 消息（非阻塞；队列满则丢弃并返回错误）。
func (h *AgentHub) Send(nodeID int64, msg Message) (err error) {
	h.mu.RLock()
	e, ok := h.agents[nodeID]
	h.mu.RUnlock()
	if !ok || e == nil {
		return ErrAgentOffline
	}
	defer func() {
		if recover() != nil {
			err = ErrAgentOffline
		}
	}()
	select {
	case e.send <- msg:
		return nil
	default:
		return errors.New("agent send queue full")
	}
}

// EnqueueAgentSend 非阻塞向下发队列写入；chan 已关闭时静默丢弃（不 panic）。
func EnqueueAgentSend(send chan<- Message, msg Message) {
	defer func() { recover() }()
	select {
	case send <- msg:
	default:
	}
}

// Online 若节点当前有已登记连接则返回 true。
func (h *AgentHub) Online(nodeID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.agents[nodeID]
	return ok
}

// WritePump 从 send 读取并写入 conn，应在独立 goroutine 中运行；send 关闭时退出。
func WritePump(conn *websocket.Conn, send <-chan Message) {
	for msg := range send {
		if err := conn.WriteJSON(msg); err != nil {
			return
		}
	}
}
