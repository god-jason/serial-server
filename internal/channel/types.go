package channel

// ChannelType 通道类型枚举
type ChannelType string

const (
	ChannelTypeTCPClient ChannelType = "tcp_client" // TCP客户端类型
	ChannelTypeTCPServer ChannelType = "tcp_server" // TCP服务端类型
	ChannelTypeMQTT      ChannelType = "mqtt"      // MQTT类型
	ChannelTypeHTTP      ChannelType = "http"      // HTTP类型
)

// Channel 通道接口定义
type Channel interface {
	Open() error                          // 打开通道
	Close() error                         // 关闭通道
	Send(data []byte) error               // 发送数据
	SetOnData(callback func([]byte))      // 设置数据接收回调
	IsConnected() bool                    // 检查是否连接
	GetConfig() ChannelConfig             // 获取配置
	GetID() string                        // 获取通道ID
}

// ChannelConfig 通道配置结构体
type ChannelConfig struct {
	ID          string      `yaml:"id"`           // 通道唯一标识
	Name        string      `yaml:"name"`         // 通道名称
	Type        ChannelType `yaml:"type"`         // 通道类型
	Enabled     bool        `yaml:"enabled"`      // 是否启用
	SerialPort  string      `yaml:"serial_port"`  // 关联串口ID
	AutoReconnect bool      `yaml:"auto_reconnect"` // 自动重连
	ReconnectInterval int    `yaml:"reconnect_interval"` // 重连间隔(秒)
	RegisterPacket string   `yaml:"register_packet"` // 注册包(十六进制)
	RegisterInterval int    `yaml:"register_interval"` // 注册间隔(秒)
	HeartbeatPacket string  `yaml:"heartbeat_packet"` // 心跳包(十六进制)
	HeartbeatInterval int   `yaml:"heartbeat_interval"` // 心跳间隔(秒)
	BufferSize int          `yaml:"buffer_size"` // 缓冲区大小

	TCPClient TCPConfig    `yaml:"tcp_client"` // TCP客户端配置
	TCPServer TCPConfig    `yaml:"tcp_server"` // TCP服务端配置
	MQTT      MQTTConfig   `yaml:"mqtt"`      // MQTT配置
	HTTP      HTTPConfig   `yaml:"http"`      // HTTP配置
}

// TCPConfig TCP配置结构体
type TCPConfig struct {
	Host string `yaml:"host"` // 主机地址
	Port int    `yaml:"port"` // 端口号
}

// MQTTConfig MQTT配置结构体
type MQTTConfig struct {
	Broker      string `yaml:"broker"`       // Broker地址
	Port        int    `yaml:"port"`         // Broker端口
	Username    string `yaml:"username"`     // 用户名
	Password    string `yaml:"password"`     // 密码
	ClientID    string `yaml:"client_id"`    // 客户端ID
	CAFile      string `yaml:"ca_file"`      // CA证书文件
	CertFile    string `yaml:"cert_file"`    // 客户端证书文件
	KeyFile     string `yaml:"key_file"`     // 客户端密钥文件
	SubscribeTopic string `yaml:"subscribe_topic"` // 订阅主题
	SendTopic      string `yaml:"send_topic"`      // 发送主题
	RegisterTopic  string `yaml:"register_topic"`  // 注册主题
	WillTopic      string `yaml:"will_topic"`      // 遗嘱主题
	WillPayload    string `yaml:"will_payload"`    // 遗嘱消息
	QOS           int    `yaml:"qos"`           // QOS级别
	KeepAlive     int    `yaml:"keep_alive"`    // 保活时间(秒)
}

// HTTPConfig HTTP配置结构体
type HTTPConfig struct {
	URL    string `yaml:"url"`    // 请求URL
	Method string `yaml:"method"` // 请求方法
	Token  string `yaml:"token"`  // 认证Token
	ContentType string `yaml:"content_type"` // 内容类型
}

// ChannelFactory 通道工厂接口
type ChannelFactory interface {
	Create(config ChannelConfig) (Channel, error) // 创建通道实例
}

var factories = make(map[ChannelType]ChannelFactory) // 通道工厂注册表

// RegisterFactory 注册通道工厂
func RegisterFactory(typ ChannelType, factory ChannelFactory) {
	factories[typ] = factory
}

// CreateChannel 根据配置创建通道
func CreateChannel(config ChannelConfig) (Channel, error) {
	factory, ok := factories[config.Type]
	if !ok {
		return nil, ErrUnknownChannelType
	}
	return factory.Create(config)
}

// ErrUnknownChannelType 未知通道类型错误
var ErrUnknownChannelType = &Error{Code: "unknown_channel_type", Message: "未知通道类型"}

// Error 通道错误类型
type Error struct {
	Code    string // 错误代码
	Message string // 错误消息
}

// Error 返回错误消息
func (e *Error) Error() string {
	return e.Message
}
