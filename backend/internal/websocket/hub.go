package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// 管理当前服务器中的所有 WebSocket 连接。
type Hub struct {
	mu      sync.RWMutex                //map 不是并发安全类型，只要满足下面任意一条，不加锁一定会出现并发读写 panic：多 goroutine 同时写 map，一个协程写另一个协程读，纯多协程只读，无任何写操作（可以不加锁），单 goroutine 操作 map（不需要锁）
	clients map[uint64]map[*Client]bool //因为一个用户可能多端登录。
}

// clients
// ├── 用户1001
// │   ├── Client A
// │   └── Client B
// ├── 用户1002
// │   └── Client C
// └── 用户1003
//
//	├── Client D
//	├── Client E
//	└── Client F
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

// 移除某个具体连接。
func (h *Hub) Unregister(client *Client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	userClients, ok := h.clients[client.UserID]
	if !ok {
		return true //这个用户在当前服务器上已经没有连接
	}

	if _, exists := userClients[client]; exists {
		delete(userClients, client)
		close(client.Send)
	}

	if len(userClients) == 0 {
		delete(h.clients, client.UserID)
		return true
	}

	return false //这个用户还有其他连接
}

func (h *Hub) SendToUser(userID uint64, msg OutgoingMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock() //拿的是读锁，读互斥
	defer h.mu.RUnlock()

	userClients, ok := h.clients[userID]
	if !ok {
		return
	}
	// 遍历用户的所有连接，发送消息，这样可以实现多端同步。
	for client := range userClients {
		select {
		case client.Send <- data:
		default: //如果 Send 队列已经满了，关闭客户端的发送通道，删除连接。
			log.Printf(
				"websocket send queue full, user_id=%d",
				userID,
			)
		}
	}
}
