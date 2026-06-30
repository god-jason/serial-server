package channel

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

// MQTTChannel MQTT通道结构体
type MQTTChannel struct {
	mutex           sync.RWMutex  // 读写锁
	config          ChannelConfig // 通道配置
	client          mqtt.Client   // MQTT客户端
	running         bool          // 是否运行中
	onData          func([]byte)  // 数据接收回调
	heartbeatTicker *time.Ticker  // 心跳定时器
	registerTicker  *time.Ticker  // 注册包定时器
	reconnectTimer  *time.Timer   // 重连定时器
}

// 初始化时注册MQTT工厂
func init() {
	RegisterFactory(ChannelTypeMQTT, &MQTTFactory{})
}

// MQTTFactory MQTT工厂
type MQTTFactory struct{}

// Create 创建MQTT通道实例
func (f *MQTTFactory) Create(config ChannelConfig) (Channel, error) {
	return &MQTTChannel{
		config: config,
	}, nil
}

// Open 打开MQTT通道
func (c *MQTTChannel) Open() error {
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

	if c.config.HeartbeatInterval > 0 {
		c.heartbeatTicker = time.NewTicker(time.Duration(c.config.HeartbeatInterval) * time.Second)
		go c.heartbeatLoop()
	}

	if c.config.RegisterInterval > 0 {
		c.registerTicker = time.NewTicker(time.Duration(c.config.RegisterInterval) * time.Second)
		go c.registerLoop()
	}

	logrus.Info("MQTT通道已打开: ", c.config.ID)

	return nil
}

// connect 连接MQTT Broker
func (c *MQTTChannel) connect() error {
	opts := mqtt.NewClientOptions()
	protocol := "tcp"
	if c.config.MQTT.CAFile != "" {
		protocol = "ssl"
		tlsConfig, err := c.createTLSConfig()
		if err != nil {
			logrus.Error("创建TLS配置失败: ", err)
			return err
		}
		opts.SetTLSConfig(tlsConfig)
	}

	brokerURL := fmt.Sprintf("%s://%s:%d", protocol, c.config.MQTT.Broker, c.config.MQTT.Port)
	opts.AddBroker(brokerURL)

	if c.config.MQTT.ClientID != "" {
		opts.SetClientID(c.config.MQTT.ClientID)
	}
	if c.config.MQTT.Username != "" {
		opts.SetUsername(c.config.MQTT.Username)
	}
	if c.config.MQTT.Password != "" {
		opts.SetPassword(c.config.MQTT.Password)
	}

	keepAlive := c.config.MQTT.KeepAlive
	if keepAlive <= 0 {
		keepAlive = 60
	}
	opts.SetKeepAlive(time.Duration(keepAlive) * time.Second)

	if c.config.MQTT.WillTopic != "" {
		opts.SetWill(c.config.MQTT.WillTopic, c.config.MQTT.WillPayload, c.config.MQTT.QOS, false)
	}

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		logrus.Info("MQTT连接成功")
		if c.config.MQTT.SubscribeTopic != "" {
			token := client.Subscribe(c.config.MQTT.SubscribeTopic, byte(c.config.MQTT.QOS), c.onMessageReceived)
			token.Wait()
			logrus.Info("已订阅主题: ", c.config.MQTT.SubscribeTopic)
		}

		if c.config.MQTT.RegisterTopic != "" && c.config.MQTT.RegisterPacket != "" {
			client.Publish(c.config.MQTT.RegisterTopic, byte(c.config.MQTT.QOS), false, hexToBytes(c.config.RegisterPacket))
		}
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logrus.Error("MQTT连接断开: ", err)
		if c.config.AutoReconnect {
			c.Close()
			c.scheduleReconnect()
		}
	})

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.WaitTimeout(10*time.Second) && token.Error() != nil {
		logrus.Error("MQTT连接失败: ", token.Error())
		return token.Error()
	}

	c.client = client

	return nil
}

// createTLSConfig 创建TLS配置
func (c *MQTTChannel) createTLSConfig() (*tls.Config, error) {
	certpool := x509.NewCertPool()
	ca, err := os.ReadFile(c.config.MQTT.CAFile)
	if err != nil {
		logrus.Error("读取CA证书失败: ", err)
		return nil, err
	}
	certpool.AppendCertsFromPEM(ca)

	cert, err := tls.LoadX509KeyPair(c.config.MQTT.CertFile, c.config.MQTT.KeyFile)
	if err != nil {
		logrus.Error("加载客户端证书失败: ", err)
		return nil, err
	}

	return &tls.Config{
		RootCAs:      certpool,
		Certificates: []tls.Certificate{cert},
	}, nil
}

// scheduleReconnect 安排重连任务
func (c *MQTTChannel) scheduleReconnect() {
	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
	}
	interval := c.config.ReconnectInterval
	if interval <= 0 {
		interval = 5
	}
	c.reconnectTimer = time.AfterFunc(time.Duration(interval)*time.Second, func() {
		logrus.Info("正在重新连接MQTT通道: ", c.config.ID)
		if err := c.connect(); err == nil {
			logrus.Info("MQTT通道重新连接成功: ", c.config.ID)
		} else {
			logrus.Error("重新连接失败: ", err)
			c.scheduleReconnect()
		}
	})
}

// Close 关闭MQTT通道
func (c *MQTTChannel) Close() error {
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

	if c.client != nil {
		c.client.Disconnect(250)
		c.client = nil
		logrus.Info("MQTT通道已关闭: ", c.config.ID)
	}

	return nil
}

// Send 发送数据到MQTT主题
func (c *MQTTChannel) Send(data []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.client == nil || !c.client.IsConnected() {
		return ErrNotConnected
	}

	if c.config.MQTT.RegisterTopic != "" && c.config.MQTT.RegisterPacket != "" {
		c.client.Publish(c.config.MQTT.RegisterTopic, byte(c.config.MQTT.QOS), false, hexToBytes(c.config.RegisterPacket))
	}

	if c.config.MQTT.SendTopic != "" {
		token := c.client.Publish(c.config.MQTT.SendTopic, byte(c.config.MQTT.QOS), false, data)
		token.Wait()
		if token.Error() != nil {
			logrus.Error("MQTT发布失败: ", token.Error())
			return token.Error()
		}
		logrus.Debug("MQTT发布到主题 ", c.config.MQTT.SendTopic, ": ", hex.EncodeToString(data))
	}

	return nil
}

// SetOnData 设置数据接收回调
func (c *MQTTChannel) SetOnData(callback func([]byte)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.onData = callback
}

// IsConnected 检查是否已连接
func (c *MQTTChannel) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.client != nil && c.client.IsConnected()
}

// GetConfig 获取通道配置
func (c *MQTTChannel) GetConfig() ChannelConfig {
	return c.config
}

// GetID 获取通道ID
func (c *MQTTChannel) GetID() string {
	return c.config.ID
}

// onMessageReceived MQTT消息接收处理函数
func (c *MQTTChannel) onMessageReceived(client mqtt.Client, msg mqtt.Message) {
	logrus.Debug("MQTT从主题 ", msg.Topic(), " 接收数据: ", hex.EncodeToString(msg.Payload()))

	if c.onData != nil {
		c.onData(msg.Payload())
	}
}

// heartbeatLoop 心跳循环
func (c *MQTTChannel) heartbeatLoop() {
	for c.running {
		select {
		case <-c.heartbeatTicker.C:
			if c.config.HeartbeatPacket != "" && c.config.MQTT.SendTopic != "" {
				c.Send(hexToBytes(c.config.HeartbeatPacket))
			}
		}
	}
}

// registerLoop 注册包发送循环
func (c *MQTTChannel) registerLoop() {
	for c.running {
		select {
		case <-c.registerTicker.C:
			if c.config.MQTT.RegisterTopic != "" && c.config.MQTT.RegisterPacket != "" {
				c.client.Publish(c.config.MQTT.RegisterTopic, byte(c.config.MQTT.QOS), false, hexToBytes(c.config.RegisterPacket))
			}
		}
	}
}
