package channel

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
)

// HTTPChannel HTTP通道结构体
type HTTPChannel struct {
	mutex   sync.RWMutex  // 读写锁
	config  ChannelConfig // 通道配置
	running bool          // 是否运行中
	onData  func([]byte)  // 数据接收回调
	client  *http.Client  // HTTP客户端
}

// 初始化时注册HTTP工厂
func init() {
	RegisterFactory(ChannelTypeHTTP, &HTTPFactory{})
}

// HTTPFactory HTTP工厂
type HTTPFactory struct{}

// Create 创建HTTP通道实例
func (f *HTTPFactory) Create(config ChannelConfig) (Channel, error) {
	return &HTTPChannel{
		config: config,
		client: &http.Client{},
	}, nil
}

// Open 打开HTTP通道
func (c *HTTPChannel) Open() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.running = true

	if c.config.RegisterPacket != "" {
		c.sendRegisterPacket()
	}

	logrus.Info("HTTP通道已打开: ", c.config.ID)

	return nil
}

// Close 关闭HTTP通道
func (c *HTTPChannel) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.running = false
	logrus.Info("HTTP通道已关闭: ", c.config.ID)

	return nil
}

// Send 发送数据通过HTTP请求
func (c *HTTPChannel) Send(data []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.running {
		return ErrNotConnected
	}

	if c.config.RegisterPacket != "" {
		c.sendRegisterPacket()
	}

	method := c.config.HTTP.Method
	if method == "" {
		method = "POST"
	}

	contentType := c.config.HTTP.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	req, err := http.NewRequest(method, c.config.HTTP.URL, bytes.NewReader(data))
	if err != nil {
		logrus.Error("创建HTTP请求失败: ", err)
		return err
	}

	if c.config.HTTP.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.HTTP.Token)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := c.client.Do(req)
	if err != nil {
		logrus.Error("HTTP请求失败: ", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Error("读取HTTP响应失败: ", err)
		return err
	}

	if resp.StatusCode >= 400 {
		logrus.Error("HTTP请求返回错误: ", resp.StatusCode, string(body))
	}

	if c.onData != nil && len(body) > 0 {
		c.onData(body)
	}

	logrus.Debug("HTTP发送数据: ", hex.EncodeToString(data))

	return nil
}

// sendRegisterPacket 发送注册包
func (c *HTTPChannel) sendRegisterPacket() error {
	if c.config.HTTP.URL == "" {
		return nil
	}

	data := hexToBytes(c.config.RegisterPacket)
	if len(data) == 0 {
		return nil
	}

	method := c.config.HTTP.Method
	if method == "" {
		method = "POST"
	}

	req, err := http.NewRequest(method, c.config.HTTP.URL, bytes.NewReader(data))
	if err != nil {
		logrus.Error("创建注册包HTTP请求失败: ", err)
		return err
	}

	if c.config.HTTP.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.HTTP.Token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		logrus.Error("发送注册包失败: ", err)
		return err
	}
	resp.Body.Close()

	return nil
}

// SetOnData 设置数据接收回调
func (c *HTTPChannel) SetOnData(callback func([]byte)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.onData = callback
}

// IsConnected 检查是否运行中
func (c *HTTPChannel) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.running
}

// GetConfig 获取通道配置
func (c *HTTPChannel) GetConfig() ChannelConfig {
	return c.config
}

// GetID 获取通道ID
func (c *HTTPChannel) GetID() string {
	return c.config.ID
}
