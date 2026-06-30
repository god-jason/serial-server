package channel

import (
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// TCPServerChannel TCP服务端通道结构体
type TCPServerChannel struct {
	mutex           sync.RWMutex     // 读写锁
	config          ChannelConfig    // 通道配置
	listener        net.Listener     // TCP监听器
	running         bool             // 是否运行中
	onData          func([]byte)     // 数据接收回调
	heartbeatTicker *time.Ticker     // 心跳定时器
	registerTicker  *time.Ticker     // 注册包定时器
	connections     []net.Conn       // 客户端连接列表
	buffer          []byte           // 接收缓冲区
}

// 初始化时注册TCP服务端工厂
func init() {
	RegisterFactory(ChannelTypeTCPServer, &TCPServerFactory{})
}

// TCPServerFactory TCP服务端工厂
type TCPServerFactory struct{}

// Create 创建TCP服务端通道实例
func (f *TCPServerFactory) Create(config ChannelConfig) (Channel, error) {
	return &TCPServerChannel{
		config: config,
		buffer: make([]byte, config.BufferSize),
	}, nil
}

// Open 打开TCP服务端通道
func (c *TCPServerChannel) Open() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.running {
		return nil
	}

	addr := fmt.Sprintf(":%d", c.config.TCPServer.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logrus.Error("TCP监听失败: ", err)
		return err
	}

	c.listener = listener
	c.running = true

	go c.acceptLoop()

	if c.config.HeartbeatInterval > 0 {
		c.heartbeatTicker = time.NewTicker(time.Duration(c.config.HeartbeatInterval) * time.Second)
		go c.heartbeatLoop()
	}

	if c.config.RegisterInterval > 0 {
		c.registerTicker = time.NewTicker(time.Duration(c.config.RegisterInterval) * time.Second)
		go c.registerLoop()
	}

	logrus.Info("TCP服务端通道已在端口 ", c.config.TCPServer.Port, " 上启动")

	return nil
}

// acceptLoop 接受连接循环
func (c *TCPServerChannel) acceptLoop() {
	for c.running {
		conn, err := c.listener.Accept()
		if err != nil {
			if c.running {
				logrus.Error("TCP接受连接失败: ", err)
			}
			break
		}

		logrus.Info("新的TCP连接来自: ", conn.RemoteAddr())

		c.mutex.Lock()
		c.connections = append(c.connections, conn)
		c.mutex.Unlock()

		if c.config.RegisterPacket != "" {
			if _, err := conn.Write(hexToBytes(c.config.RegisterPacket)); err != nil {
				logrus.Error("发送注册包失败: ", err)
				conn.Close()
				continue
			}
		}

		go c.handleConnection(conn)
	}
}

// handleConnection 处理单个客户端连接
func (c *TCPServerChannel) handleConnection(conn net.Conn) {
	for c.running {
		n, err := conn.Read(c.buffer)
		if err != nil {
			logrus.Info("TCP连接已关闭: ", conn.RemoteAddr())
			c.removeConnection(conn)
			break
		}

		if n > 0 {
			data := make([]byte, n)
			copy(data, c.buffer[:n])

			if c.onData != nil {
				c.onData(data)
			}

			logrus.Debug("TCP服务端从 ", conn.RemoteAddr(), " 接收数据: ", hex.EncodeToString(data))
		}
	}
}

// removeConnection 移除断开的连接
func (c *TCPServerChannel) removeConnection(conn net.Conn) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i, connItem := range c.connections {
		if connItem == conn {
			c.connections = append(c.connections[:i], c.connections[i+1:]...)
			break
		}
	}
}

// Close 关闭TCP服务端通道
func (c *TCPServerChannel) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.running = false

	if c.heartbeatTicker != nil {
		c.heartbeatTicker.Stop()
	}
	if c.registerTicker != nil {
		c.registerTicker.Stop()
	}

	for _, conn := range c.connections {
		conn.Close()
	}
	c.connections = nil

	if c.listener != nil {
		err := c.listener.Close()
		c.listener = nil
		logrus.Info("TCP服务端通道已关闭")
		return err
	}

	return nil
}

// Send 向所有连接的客户端发送数据（数据多发）
func (c *TCPServerChannel) Send(data []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.connections) == 0 {
		return ErrNoConnections
	}

	for _, conn := range c.connections {
		if c.config.RegisterPacket != "" {
			if _, err := conn.Write(hexToBytes(c.config.RegisterPacket)); err != nil {
				logrus.Error("发送注册包失败: ", err)
				continue
			}
		}

		_, err := conn.Write(data)
		if err != nil {
			logrus.Error("TCP发送数据失败: ", err)
		} else {
			logrus.Debug("TCP服务端向 ", conn.RemoteAddr(), " 发送数据: ", hex.EncodeToString(data))
		}
	}

	return nil
}

// SetOnData 设置数据接收回调
func (c *TCPServerChannel) SetOnData(callback func([]byte)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.onData = callback
}

// IsConnected 检查是否有客户端连接
func (c *TCPServerChannel) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.connections) > 0
}

// GetConfig 获取通道配置
func (c *TCPServerChannel) GetConfig() ChannelConfig {
	return c.config
}

// GetID 获取通道ID
func (c *TCPServerChannel) GetID() string {
	return c.config.ID
}

// heartbeatLoop 心跳循环
func (c *TCPServerChannel) heartbeatLoop() {
	for c.running {
		select {
		case <-c.heartbeatTicker.C:
			if c.config.HeartbeatPacket != "" {
				c.Send(hexToBytes(c.config.HeartbeatPacket))
			}
		}
	}
}

// registerLoop 注册包发送循环
func (c *TCPServerChannel) registerLoop() {
	for c.running {
		select {
		case <-c.registerTicker.C:
			c.mutex.Lock()
			for _, conn := range c.connections {
				if c.config.RegisterPacket != "" {
					conn.Write(hexToBytes(c.config.RegisterPacket))
				}
			}
			c.mutex.Unlock()
		}
	}
}

// ErrNoConnections 无连接错误
var ErrNoConnections = &Error{Code: "no_connections", Message: "没有连接的客户端"}
