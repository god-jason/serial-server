package channel

import (
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// TCPClientChannel TCP客户端通道结构体
type TCPClientChannel struct {
	mutex           sync.RWMutex     // 读写锁
	config          ChannelConfig    // 通道配置
	conn            net.Conn         // TCP连接
	running         bool             // 是否运行中
	onData          func([]byte)     // 数据接收回调
	heartbeatTicker *time.Ticker     // 心跳定时器
	registerTicker  *time.Ticker     // 注册包定时器
	reconnectTimer  *time.Timer      // 重连定时器
	buffer          []byte           // 接收缓冲区
}

// 初始化时注册TCP客户端工厂
func init() {
	RegisterFactory(ChannelTypeTCPClient, &TCPClientFactory{})
}

// TCPClientFactory TCP客户端工厂
type TCPClientFactory struct{}

// Create 创建TCP客户端通道实例
func (f *TCPClientFactory) Create(config ChannelConfig) (Channel, error) {
	return &TCPClientChannel{
		config: config,
		buffer: make([]byte, config.BufferSize),
	}, nil
}

// Open 打开TCP客户端通道
func (c *TCPClientChannel) Open() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.running {
		return nil
	}

	if err := c.connect(); err != nil {
		if c.config.AutoReconnect {
			c.scheduleReconnect()
		}
		return err
	}

	c.running = true
	go c.readLoop()

	if c.config.HeartbeatInterval > 0 {
		c.heartbeatTicker = time.NewTicker(time.Duration(c.config.HeartbeatInterval) * time.Second)
		go c.heartbeatLoop()
	}

	if c.config.RegisterInterval > 0 {
		c.registerTicker = time.NewTicker(time.Duration(c.config.RegisterInterval) * time.Second)
		go c.registerLoop()
	}

	logrus.Info("TCP客户端通道已打开: ", c.config.ID)

	return nil
}

// connect 建立TCP连接
func (c *TCPClientChannel) connect() error {
	addr := fmt.Sprintf("%s:%d", c.config.TCPClient.Host, c.config.TCPClient.Port)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		logrus.Error("TCP连接失败: ", err)
		return err
	}

	c.conn = conn

	if c.config.RegisterPacket != "" {
		if err := c.sendRegisterPacket(); err != nil {
			conn.Close()
			return err
		}
	}

	return nil
}

// scheduleReconnect 安排重连任务
func (c *TCPClientChannel) scheduleReconnect() {
	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
	}
	interval := c.config.ReconnectInterval
	if interval <= 0 {
		interval = 5
	}
	c.reconnectTimer = time.AfterFunc(time.Duration(interval)*time.Second, func() {
		logrus.Info("正在重新连接TCP客户端通道: ", c.config.ID)
		if err := c.connect(); err == nil {
			logrus.Info("TCP客户端通道重新连接成功: ", c.config.ID)
			go c.readLoop()
		} else {
			logrus.Error("重新连接失败: ", err)
			c.scheduleReconnect()
		}
	})
}

// Close 关闭TCP客户端通道
func (c *TCPClientChannel) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.running = false

	if c.heartbeatTicker != nil {
		c.heartbeatTicker.Stop()
	}
	if c.registerTicker != nil {
		c.registerTicker.Stop()
	}
	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
	}

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		logrus.Info("TCP客户端通道已关闭: ", c.config.ID)
		return err
	}

	return nil
}

// Send 发送数据
func (c *TCPClientChannel) Send(data []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.running || c.conn == nil {
		return ErrNotConnected
	}

	if c.config.RegisterPacket != "" {
		if _, err := c.conn.Write(hexToBytes(c.config.RegisterPacket)); err != nil {
			logrus.Error("发送注册包失败: ", err)
			return err
		}
	}

	_, err := c.conn.Write(data)
	if err != nil {
		logrus.Error("TCP发送数据失败: ", err)
		if c.config.AutoReconnect {
			c.Close()
			c.scheduleReconnect()
		}
		return err
	}

	logrus.Debug("TCP客户端发送数据: ", hex.EncodeToString(data))

	return nil
}

// SetOnData 设置数据接收回调
func (c *TCPClientChannel) SetOnData(callback func([]byte)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.onData = callback
}

// IsConnected 检查是否已连接
func (c *TCPClientChannel) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.conn != nil
}

// GetConfig 获取通道配置
func (c *TCPClientChannel) GetConfig() ChannelConfig {
	return c.config
}

// GetID 获取通道ID
func (c *TCPClientChannel) GetID() string {
	return c.config.ID
}

// readLoop 数据读取循环
func (c *TCPClientChannel) readLoop() {
	for c.running {
		n, err := c.conn.Read(c.buffer)
		if err != nil {
			if c.running {
				logrus.Error("TCP读取数据失败: ", err)
				if c.config.AutoReconnect {
					c.Close()
					c.scheduleReconnect()
				}
			}
			break
		}

		if n > 0 {
			data := make([]byte, n)
			copy(data, c.buffer[:n])

			if c.onData != nil {
				c.onData(data)
			}

			logrus.Debug("TCP客户端接收数据: ", hex.EncodeToString(data))
		}
	}
}

// heartbeatLoop 心跳循环
func (c *TCPClientChannel) heartbeatLoop() {
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
func (c *TCPClientChannel) registerLoop() {
	for c.running {
		select {
		case <-c.registerTicker.C:
			c.sendRegisterPacket()
		}
	}
}

// sendRegisterPacket 发送注册包
func (c *TCPClientChannel) sendRegisterPacket() error {
	if c.conn == nil {
		return ErrNotConnected
	}
	data := hexToBytes(c.config.RegisterPacket)
	_, err := c.conn.Write(data)
	if err != nil {
		logrus.Error("发送注册包失败: ", err)
		return err
	}
	return nil
}

// hexToBytes 将十六进制字符串转换为字节数组
func hexToBytes(s string) []byte {
	if s == "" {
		return nil
	}
	data, _ := hex.DecodeString(s)
	return data
}

// ErrNotConnected 未连接错误
var ErrNotConnected = &Error{Code: "not_connected", Message: "通道未连接"}
