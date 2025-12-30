package websocket

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Hub WebSocket 连接中心
type Hub struct {
	clients    map[uint]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     *zap.Logger
}

// NewHub 创建 Hub
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:    make(map[uint]*Client),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Run 运行 Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			h.logger.Info("客户端已连接", zap.Uint("user_id", client.UserID))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()
			h.logger.Info("客户端已断开", zap.Uint("user_id", client.UserID))

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.UserID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast 广播消息
func (h *Hub) Broadcast(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("序列化消息失败", zap.Error(err))
		return
	}
	h.broadcast <- data
}

// SendToUser 发送消息给指定用户
func (h *Hub) SendToUser(userID uint, message interface{}) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("序列化消息失败", zap.Error(err))
		return
	}

	select {
	case client.Send <- data:
	default:
		close(client.Send)
		h.mu.Lock()
		delete(h.clients, userID)
		h.mu.Unlock()
	}
}

// Client WebSocket 客户端
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   uint
	Username string
}

// ReadPump 读取消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.Error("WebSocket 错误", zap.Error(err))
			}
			break
		}

		// 处理消息
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			c.Hub.logger.Error("解析消息失败", zap.Error(err))
			continue
		}

		// 这里可以添加消息处理逻辑
		c.Hub.logger.Info("收到消息", zap.Any("message", msg))
	}
}

// WritePump 写入消息
func (c *Client) WritePump() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.Hub.logger.Error("写入消息失败", zap.Error(err))
				return
			}
		}
	}
}

