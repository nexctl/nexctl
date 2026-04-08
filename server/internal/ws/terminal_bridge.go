package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

const browserTerminalQueue = 128

// TerminalBridge 将 Agent 上行的 terminal_output 转发到对应浏览器会话。
type TerminalBridge struct {
	mu       sync.RWMutex
	sessions map[string]*browserTerminalSession
}

type browserTerminalSession struct {
	NodeID    int64
	SessionID string
	ToBrowser chan Message
}

func NewTerminalBridge() *TerminalBridge {
	return &TerminalBridge{sessions: make(map[string]*browserTerminalSession)}
}

func (b *TerminalBridge) Register(sessionID string, nodeID int64) (toBrowser chan Message, unregister func()) {
	ch := make(chan Message, browserTerminalQueue)
	s := &browserTerminalSession{NodeID: nodeID, SessionID: sessionID, ToBrowser: ch}
	b.mu.Lock()
	b.sessions[sessionID] = s
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		if cur, ok := b.sessions[sessionID]; ok && cur == s {
			delete(b.sessions, sessionID)
		}
		b.mu.Unlock()
		close(ch)
	}
}

// DispatchFromAgent 将 Agent 发来的终端输出转发到浏览器会话。
func (b *TerminalBridge) DispatchFromAgent(sessionID string, msg Message) bool {
	b.mu.RLock()
	s, ok := b.sessions[sessionID]
	b.mu.RUnlock()
	if !ok || s == nil {
		return false
	}
	select {
	case s.ToBrowser <- msg:
		return true
	default:
		return false
	}
}

// BrowserWritePump 将 ToBrowser 中的消息写入浏览器 WebSocket。
func BrowserWritePump(conn *websocket.Conn, toBrowser <-chan Message) {
	for msg := range toBrowser {
		if err := conn.WriteJSON(msg); err != nil {
			return
		}
	}
}
