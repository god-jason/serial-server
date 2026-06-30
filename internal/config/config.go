package config

import (
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Config 全局配置结构体
type Config struct {
	Gateway  GatewayConfig   `yaml:"gateway" json:"gateway"`   // 网关网络配置
	Serial   SerialConfig    `yaml:"serial" json:"serial"`     // 串口配置
	Channels []ChannelConfig `yaml:"channels" json:"channels"` // 通道配置列表
	System   SystemConfig    `yaml:"system" json:"system"`     // 系统配置
}

// GatewayConfig 网关网络配置
type GatewayConfig struct {
	IP      string `yaml:"ip" json:"ip"`           // IP地址
	Netmask string `yaml:"netmask" json:"netmask"` // 子网掩码
	Gateway string `yaml:"gateway" json:"gateway"` // 默认网关
	DNS     string `yaml:"dns" json:"dns"`         // DNS服务器
	DHCP    bool   `yaml:"dhcp" json:"dhcp"`       // 是否启用DHCP
}

// SerialConfig 串口配置
type SerialConfig struct {
	Ports []SerialPortConfig `yaml:"ports" json:"ports"` // 串口列表
}

// SerialPortConfig 单个串口配置
type SerialPortConfig struct {
	ID             string `yaml:"id" json:"id"`                           // 串口唯一标识
	Name           string `yaml:"name" json:"name"`                       // 串口名称
	Port           string `yaml:"port" json:"port"`                       // 串口设备路径
	BaudRate       int    `yaml:"baud_rate" json:"baud_rate"`             // 波特率
	DataBits       int    `yaml:"data_bits" json:"data_bits"`             // 数据位
	Parity         string `yaml:"parity" json:"parity"`                   // 校验位
	StopBits       int    `yaml:"stop_bits" json:"stop_bits"`             // 停止位
	FlowControl    string `yaml:"flow_control" json:"flow_control"`       // 流控方式
	DelayPackaging int    `yaml:"delay_packaging" json:"delay_packaging"` // 延迟封包时间(毫秒)
	DelayTimeout   int    `yaml:"delay_timeout" json:"delay_timeout"`     // 封包超时时间(毫秒)
	Protocol       string `yaml:"protocol" json:"protocol"`               // 协议类型
	Enabled        bool   `yaml:"enabled" json:"enabled"`                 // 是否启用
}

// ChannelConfig 通道配置
type ChannelConfig struct {
	ID                string     `yaml:"id" json:"id"`                 // 通道唯一标识
	Name              string     `yaml:"name" json:"name"`               // 通道名称
	Type              string     `yaml:"type" json:"type"`               // 通道类型
	Enabled           bool       `yaml:"enabled" json:"enabled"`            // 是否启用
	SerialPort        string     `yaml:"serial_port" json:"serial_port"`        // 关联串口ID
	AutoReconnect     bool       `yaml:"auto_reconnect" json:"auto_reconnect"`     // 自动重连
	ReconnectInterval int        `yaml:"reconnect_interval" json:"reconnect_interval"` // 重连间隔(秒)
	RegisterPacket    string     `yaml:"register_packet" json:"register_packet"`    // 注册包(十六进制)
	RegisterInterval  int        `yaml:"register_interval" json:"register_interval"`  // 注册间隔(秒)
	HeartbeatPacket   string     `yaml:"heartbeat_packet" json:"heartbeat_packet"`   // 心跳包(十六进制)
	HeartbeatInterval int        `yaml:"heartbeat_interval" json:"heartbeat_interval"` // 心跳间隔(秒)
	BufferSize        int        `yaml:"buffer_size" json:"buffer_size"`        // 缓冲区大小

	TCPClient TCPConfig  `yaml:"tcp_client" json:"tcp_client"` // TCP客户端配置
	TCPServer TCPConfig  `yaml:"tcp_server" json:"tcp_server"` // TCP服务端配置
	MQTT      MQTTConfig `yaml:"mqtt" json:"mqtt"`       // MQTT配置
	HTTP      HTTPConfig `yaml:"http" json:"http"`       // HTTP配置
}

// TCPConfig TCP配置
type TCPConfig struct {
	Host string `yaml:"host" json:"host"` // 主机地址
	Port int    `yaml:"port" json:"port"` // 端口号
}

// MQTTConfig MQTT配置
type MQTTConfig struct {
	Broker         string `yaml:"broker" json:"broker"`          // Broker地址
	Port           int    `yaml:"port" json:"port"`            // Broker端口
	Username       string `yaml:"username" json:"username"`        // 用户名
	Password       string `yaml:"password" json:"password"`        // 密码
	ClientID       string `yaml:"client_id" json:"client_id"`       // 客户端ID
	CAFile         string `yaml:"ca_file" json:"ca_file"`         // CA证书文件
	CertFile       string `yaml:"cert_file" json:"cert_file"`       // 客户端证书文件
	KeyFile        string `yaml:"key_file" json:"key_file"`        // 客户端密钥文件
	SubscribeTopic string `yaml:"subscribe_topic" json:"subscribe_topic"` // 订阅主题
	SendTopic      string `yaml:"send_topic" json:"send_topic"`      // 发送主题
	RegisterTopic  string `yaml:"register_topic" json:"register_topic"`  // 注册主题
	WillTopic      string `yaml:"will_topic" json:"will_topic"`      // 遗嘱主题
	WillPayload    string `yaml:"will_payload" json:"will_payload"`    // 遗嘱消息
	QOS            int    `yaml:"qos" json:"qos"`             // QOS级别
	KeepAlive      int    `yaml:"keep_alive" json:"keep_alive"`      // 保活时间(秒)
}

// HTTPConfig HTTP配置
type HTTPConfig struct {
	URL         string `yaml:"url" json:"url"`          // 请求URL
	Method      string `yaml:"method" json:"method"`       // 请求方法
	Token       string `yaml:"token" json:"token"`        // 认证Token
	ContentType string `yaml:"content_type" json:"content_type"` // 内容类型
}

// SystemConfig 系统配置
type SystemConfig struct {
	WebPort       int            `yaml:"web_port" json:"web_port"`        // Web服务端口
	LogLevel      string         `yaml:"log_level" json:"log_level"`       // 日志级别
	LogFile       string         `yaml:"log_file" json:"log_file"`        // 日志文件路径
	LogMaxSize    int            `yaml:"log_max_size" json:"log_max_size"`    // 日志文件最大大小(MB)
	LogMaxBackups int            `yaml:"log_max_backups" json:"log_max_backups"` // 日志文件最大备份数
	Password      string         `yaml:"password" json:"password"`        // 登录密码
	OEM           OEMConfig      `yaml:"oem" json:"oem"`             // OEM配置
	WiFi          WiFiConfig     `yaml:"wifi" json:"wifi"`            // WiFi配置
	Module4G      Module4GConfig `yaml:"module_4g" json:"module_4g"`       // 4G模组配置
}

// OEMConfig OEM配置
type OEMConfig struct {
	Name       string `yaml:"name" json:"name"`        // OEM名称
	Logo       string `yaml:"logo" json:"logo"`        // Logo URL
	ThemeColor string `yaml:"theme_color" json:"theme_color"` // 主题颜色
}

// WiFiConfig WiFi配置
type WiFiConfig struct {
	SSID     string `yaml:"ssid" json:"ssid"`     // WiFi名称
	Password string `yaml:"password" json:"password"` // WiFi密码
	Enabled  bool   `yaml:"enabled" json:"enabled"`  // 是否启用
}

// Module4GConfig 4G模组配置
type Module4GConfig struct {
	APN     string `yaml:"apn" json:"apn"`     // APN接入点
	Enabled bool   `yaml:"enabled" json:"enabled"` // 是否启用
}

var (
	config     *Config         // 全局配置实例
	mutex      sync.RWMutex    // 配置读写锁
	configFile = "config.yaml" // 默认配置文件路径
)

// Load 加载配置文件
func Load(file string) error {
	if file != "" {
		configFile = file
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		logrus.Error("加载配置文件失败: ", err)
		return err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		logrus.Error("解析配置文件失败: ", err)
		return err
	}

	mutex.Lock()
	config = &cfg
	mutex.Unlock()

	logrus.Info("配置文件加载成功")

	return nil
}

// Get 获取当前配置
func Get() *Config {
	mutex.RLock()
	defer mutex.RUnlock()
	return config
}

// Save 保存配置到文件
func Save() error {
	mutex.RLock()
	cfg := config
	mutex.RUnlock()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		logrus.Error("序列化配置失败: ", err)
		return err
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		logrus.Error("保存配置文件失败: ", err)
		return err
	}

	logrus.Info("配置文件保存成功")

	return nil
}

// Update 更新配置
func Update(newConfig *Config) {
	mutex.Lock()
	config = newConfig
	mutex.Unlock()
	logrus.Info("配置已更新")
}

// Watch 监听配置文件变化
func Watch(callback func()) {
	go func() {
		prevModTime := time.Time{}
		for {
			time.Sleep(5 * time.Second)
			fi, err := os.Stat(configFile)
			if err != nil {
				continue
			}
			if fi.ModTime().After(prevModTime) {
				prevModTime = fi.ModTime()
				logrus.Info("配置文件发生变化，正在重新加载...")
				if err := Load(configFile); err != nil {
					logrus.Error("重新加载配置失败: ", err)
					continue
				}
				callback()
			}
		}
	}()
}

// GetDefault 获取默认配置
func GetDefault() *Config {
	return &Config{
		Gateway: GatewayConfig{
			DHCP: true,
		},
		Serial: SerialConfig{
			Ports: []SerialPortConfig{
				{
					ID:             "serial1",
					Name:           "COM1",
					Port:           "/dev/ttyS0",
					BaudRate:       9600,
					DataBits:       8,
					Parity:         "none",
					StopBits:       1,
					FlowControl:    "none",
					DelayPackaging: 10,
					DelayTimeout:   100,
					Protocol:       "raw",
					Enabled:        false,
				},
			},
		},
		Channels: []ChannelConfig{},
		System: SystemConfig{
			WebPort:       8080,
			LogLevel:      "info",
			LogFile:       "logs/app.log",
			LogMaxSize:    10,
			LogMaxBackups: 5,
			Password:      "admin",
			OEM: OEMConfig{
				Name:       "串口服务器",
				ThemeColor: "#409EFF",
			},
		},
	}
}
