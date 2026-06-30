package websocket

import (
	"encoding/hex"
	"net/http"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// WebSocketManager WebSocket管理器
type WebSocketManager struct {
	mutex             sync.RWMutex                 // 读写锁
	serialConnections map[string][]*websocket.Conn // 串口调试连接
	termConnections   map[string][]*websocket.Conn // 终端连接
}

// NewWebSocketManager 创建WebSocket管理器实例
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		serialConnections: make(map[string][]*websocket.Conn),
		termConnections:   make(map[string][]*websocket.Conn),
	}
}

// WebSocket升级器配置
var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// AddSerialConnection 添加串口调试连接
func (w *WebSocketManager) AddSerialConnection(portID string, conn *websocket.Conn) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.serialConnections[portID] = append(w.serialConnections[portID], conn)
}

// RemoveSerialConnection 移除串口调试连接
func (w *WebSocketManager) RemoveSerialConnection(portID string, conn *websocket.Conn) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	conns := w.serialConnections[portID]
	for i, c := range conns {
		if c == conn {
			w.serialConnections[portID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

// BroadcastSerialData 广播串口数据到所有连接的客户端
func (w *WebSocketManager) BroadcastSerialData(portID string, data []byte) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	for _, conn := range w.serialConnections[portID] {
		if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
			logrus.Error("WebSocket写入失败: ", err)
		}
	}
}

// AddTermConnection 添加终端连接
func (w *WebSocketManager) AddTermConnection(sessionID string, conn *websocket.Conn) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.termConnections[sessionID] = append(w.termConnections[sessionID], conn)
}

// RemoveTermConnection 移除终端连接
func (w *WebSocketManager) RemoveTermConnection(sessionID string, conn *websocket.Conn) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	conns := w.termConnections[sessionID]
	for i, c := range conns {
		if c == conn {
			w.termConnections[sessionID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

// StartTerminal 启动xterm终端
func (w *WebSocketManager) StartTerminal(conn *websocket.Conn) {
	var cmd *exec.Cmd
	cmd = exec.Command("bash")

	f, err := pty.Start(cmd)
	if err != nil {
		logrus.Error("启动终端失败: ", err)
		return
	}

	defer func() {
		f.Close()
		cmd.Wait()
	}()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := f.Read(buf)
			if err != nil {
				break
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				break
			}
		}
	}()

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			f.Write(message)
		}
	}()
}

// SerialDebugHandler 串口调试处理器
type SerialDebugHandler struct {
	wsManager  *WebSocketManager // WebSocket管理器
	serialPort SerialPort        // 串口接口
}

// SerialPort 串口接口定义
type SerialPort interface {
	Write(data []byte) error // 写入数据
	IsOpen() bool            // 检查是否打开
}

// NewSerialDebugHandler 创建串口调试处理器实例
func NewSerialDebugHandler(wsManager *WebSocketManager, serialPort SerialPort) *SerialDebugHandler {
	return &SerialDebugHandler{
		wsManager:  wsManager,
		serialPort: serialPort,
	}
}

// Handle 处理WebSocket串口调试消息
func (h *SerialDebugHandler) Handle(conn *websocket.Conn) {
	defer conn.Close()

	for {
		msgType, message, err := conn.ReadMessage()
		if err != nil {
			logrus.Error("WebSocket读取失败: ", err)
			break
		}

		if msgType == websocket.BinaryMessage {
			if h.serialPort != nil && h.serialPort.IsOpen() {
				h.serialPort.Write(message)
			}
		} else if msgType == websocket.TextMessage {
			data, err := hex.DecodeString(string(message))
			if err == nil && h.serialPort != nil && h.serialPort.IsOpen() {
				h.serialPort.Write(data)
			}
		}
	}
}
