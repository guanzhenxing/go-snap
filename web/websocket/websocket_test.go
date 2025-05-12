package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	handler := NewHandler()
	assert.NotNil(t, handler)
	assert.Equal(t, defaultUpgrader.ReadBufferSize, handler.upgrader.ReadBufferSize)
	assert.Equal(t, defaultUpgrader.WriteBufferSize, handler.upgrader.WriteBufferSize)
	assert.NotNil(t, handler.log)
	assert.Equal(t, 60*time.Second, handler.pingInterval)
	assert.Equal(t, int64(1024*1024), handler.maxMessageSize)
	assert.NotNil(t, handler.clients)
	assert.NotNil(t, handler.broadcast)
}

func TestHandlerOptions(t *testing.T) {
	customUpgrader := websocket.Upgrader{
		ReadBufferSize:  2048,
		WriteBufferSize: 2048,
	}
	customPingInterval := 30 * time.Second
	customMaxMessageSize := int64(2048 * 1024)
	customBufferSize := 512

	handler := NewHandler(
		WithUpgrader(customUpgrader),
		WithPingInterval(customPingInterval),
		WithMaxMessageSize(customMaxMessageSize),
		WithBroadcastBuffer(customBufferSize),
	)

	assert.Equal(t, customUpgrader, handler.upgrader)
	assert.Equal(t, customPingInterval, handler.pingInterval)
	assert.Equal(t, customMaxMessageSize, handler.maxMessageSize)
	assert.Equal(t, customBufferSize, cap(handler.broadcast))
}

func TestWebSocketConnection(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试服务器
	router := gin.New()
	handler := NewHandler()
	router.GET("/ws", handler.Middleware())

	server := httptest.NewServer(router)
	defer server.Close()

	// 将http://test-server转换为ws://test-server
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// 连接WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	// 测试客户端数量
	assert.Equal(t, 1, handler.ClientCount())

	// 测试发送消息
	testMessage := []byte("test message")
	err = ws.WriteMessage(websocket.TextMessage, testMessage)
	require.NoError(t, err)

	// 测试广播
	handler.Broadcast([]byte("broadcast message"))

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 关闭连接
	ws.Close()

	// 等待连接关闭
	time.Sleep(100 * time.Millisecond)

	// 验证客户端数量
	assert.Equal(t, 0, handler.ClientCount())
}

func TestWebSocketMiddleware(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试服务器
	router := gin.New()
	handler := NewHandler()
	router.GET("/ws", handler.Middleware())

	// 测试HTTP请求（非WebSocket）
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ws", nil)
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebSocketBroadcast(t *testing.T) {
	handler := NewHandler()

	// 创建多个测试连接
	connections := make([]*websocket.Conn, 3)
	for i := range connections {
		// 创建测试服务器
		router := gin.New()
		router.GET("/ws", handler.Middleware())
		server := httptest.NewServer(router)
		defer server.Close()

		// 连接WebSocket
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
		ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		connections[i] = ws
		defer ws.Close()
	}

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 验证客户端数量
	assert.Equal(t, 3, handler.ClientCount())

	// 广播消息
	testMessage := []byte("broadcast test")
	handler.Broadcast(testMessage)

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 关闭所有连接
	for _, conn := range connections {
		conn.Close()
	}

	// 等待连接关闭
	time.Sleep(100 * time.Millisecond)

	// 验证客户端数量
	assert.Equal(t, 0, handler.ClientCount())
}
