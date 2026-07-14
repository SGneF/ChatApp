package websocket

import (
	"encoding/json"
	"sync"
)

// 在线用户连接管理器。
type Hub struct {
	mu      sync.RWMutex
	clients map[uint64]map[*Client]bool //用户可能有多个连接：pc端，app端
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uint64]map[*Client]bool),
	}
}

// 登记连接成功的用户websocket连接
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[*Client]bool)
	}

	h.clients[client.UserID][client] = true
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	userClients, ok := h.clients[client.UserID]
	if !ok {
		return
	}

	if _, exists := userClients[client]; exists {
		delete(userClients, client)
		close(client.Send)
	}

	if len(userClients) == 0 {
		delete(h.clients, client.UserID)
	}
}

func (h *Hub) SendToUser(userID uint64, msg OutgoingMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	userClients, ok := h.clients[userID]
	if !ok {
		return
	}

	for client := range userClients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(userClients, client)
		}
	}
}

func (h *Hub) IsOnline(userID uint64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clients[userID]) > 0
}
