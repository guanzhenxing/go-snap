// Package websocket 提供WebSocket支持
package websocket

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/guanzhenxing/go-snap/errors"
	"github.com/guanzhenxing/go-snap/logger"
)

// 默认WebSocket配置
var defaultUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// Handler WebSocket处理器
type Handler struct {
	upgrader       websocket.Upgrader
	log            logger.Logger
	pingInterval   time.Duration
	maxMessageSize int64
	clients        map[*websocket.Conn]bool
	clientsMutex   sync.Mutex
	broadcast      chan []byte
}

// HandlerOption WebSocket处理器配置选项
type HandlerOption func(*Handler)

// WithUpgrader 设置自定义Upgrader
func WithUpgrader(upgrader websocket.Upgrader) HandlerOption {
	return func(h *Handler) {
		h.upgrader = upgrader
	}
}

// WithLogger 设置自定义日志器
func WithLogger(log logger.Logger) HandlerOption {
	return func(h *Handler) {
		h.log = log
	}
}

// WithPingInterval 设置Ping间隔
func WithPingInterval(interval time.Duration) HandlerOption {
	return func(h *Handler) {
		h.pingInterval = interval
	}
}

// WithMaxMessageSize 设置最大消息大小
func WithMaxMessageSize(size int64) HandlerOption {
	return func(h *Handler) {
		h.maxMessageSize = size
	}
}

// WithBroadcastBuffer 设置广播缓冲区大小
func WithBroadcastBuffer(size int) HandlerOption {
	return func(h *Handler) {
		h.broadcast = make(chan []byte, size)
	}
}

// NewHandler 创建新的WebSocket处理器
func NewHandler(opts ...HandlerOption) *Handler {
	handler := &Handler{
		upgrader:       defaultUpgrader,
		log:            logger.New(),
		pingInterval:   60 * time.Second,
		maxMessageSize: 1024 * 1024, // 1MB
		clients:        make(map[*websocket.Conn]bool),
		broadcast:      make(chan []byte, 256),
	}

	// 应用选项
	for _, opt := range opts {
		opt(handler)
	}

	// 启动广播处理协程
	go handler.handleBroadcast()

	return handler
}

// handleBroadcast 处理广播消息
func (h *Handler) handleBroadcast() {
	for message := range h.broadcast {
		h.clientsMutex.Lock()
		for client := range h.clients {
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				h.log.Error(errors.Wrap(err, "failed to broadcast message").Error())
				client.Close()
				delete(h.clients, client)
			}
		}
		h.clientsMutex.Unlock()
	}
}

// HandleConnection 处理WebSocket连接
func (h *Handler) HandleConnection(c *gin.Context) {
	// 升级HTTP连接为WebSocket连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error(errors.Wrap(err, "failed to upgrade connection").Error())
		return
	}

	// 关闭连接
	defer func() {
		h.clientsMutex.Lock()
		delete(h.clients, conn)
		h.clientsMutex.Unlock()
		conn.Close()
	}()

	// 设置读取限制
	conn.SetReadLimit(h.maxMessageSize)

	// 设置连接参数
	conn.SetReadDeadline(time.Now().Add(h.pingInterval + 10*time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(h.pingInterval + 10*time.Second))
		return nil
	})

	// 添加到客户端列表
	h.clientsMutex.Lock()
	h.clients[conn] = true
	h.clientsMutex.Unlock()

	// 启动ping协程
	stopPing := make(chan struct{})
	go h.pingClient(conn, stopPing)

	// 消息处理循环
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.log.Error(errors.Wrap(err, "unexpected close error").Error())
			}
			break
		}

		// 处理消息
		h.handleMessage(conn, message)
	}

	// 停止ping
	close(stopPing)
}

// pingClient 定期发送ping消息保持连接活跃
func (h *Handler) pingClient(conn *websocket.Conn, stop chan struct{}) {
	ticker := time.NewTicker(h.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				h.log.Error(errors.Wrap(err, "failed to send ping").Error())
				return
			}
		case <-stop:
			return
		}
	}
}

// handleMessage 处理接收到的WebSocket消息
func (h *Handler) handleMessage(conn *websocket.Conn, message []byte) {
	// 默认实现是直接广播消息
	// 实际使用时通常需要重写此方法进行自定义处理
	h.broadcast <- message
}

// Broadcast 广播消息给所有连接的客户端
func (h *Handler) Broadcast(message []byte) {
	h.broadcast <- message
}

// ClientCount 获取当前连接的客户端数量
func (h *Handler) ClientCount() int {
	h.clientsMutex.Lock()
	defer h.clientsMutex.Unlock()
	return len(h.clients)
}

// Middleware 返回用于集成到Gin路由的中间件
func (h *Handler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.HandleConnection(c)
	}
}

// NewHandlerFunc 创建一个新的WebSocket处理函数
func NewHandlerFunc(opts ...HandlerOption) gin.HandlerFunc {
	handler := NewHandler(opts...)
	return handler.Middleware()
}
