package serial

import (
	"bytes"
	"encoding/hex"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SerialPort 串口端口结构体
type SerialPort struct {
	mutex             sync.RWMutex     // 读写锁
	config            *SerialConfig    // 串口配置
	running           bool             // 是否运行中
	onData            func([]byte)     // 数据接收回调函数
	buffer            bytes.Buffer     // 延迟封包缓冲区
	delayTimer        *time.Timer      // 延迟封包定时器
	delayTimeoutTimer *time.Timer      // 封包超时定时器
	writeFn           func([]byte) (int, error) // 实际写入函数
	closeFn           func() error     // 实际关闭函数
}

// SerialConfig 串口配置结构体
type SerialConfig struct {
	ID             string // 串口唯一标识
	Name           string // 串口名称
	Port           string // 串口设备路径
	BaudRate       int    // 波特率
	DataBits       int    // 数据位
	Parity         string // 校验位
	StopBits       int    // 停止位
	FlowControl    string // 流控方式
	DelayPackaging int    // 延迟封包时间(毫秒)
	DelayTimeout   int    // 封包超时时间(毫秒)
	Protocol       string // 协议类型
	Enabled        bool   // 是否启用
}

// NewSerialPort 创建新的串口实例
func NewSerialPort(config *SerialConfig) *SerialPort {
	return &SerialPort{
		config: config,
	}
}

// Open 打开串口
func (s *SerialPort) Open() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return nil
	}

	s.running = true

	go s.readLoop()

	logrus.Info("串口已打开: ", s.config.Port)

	return nil
}

// Close 关闭串口
func (s *SerialPort) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return nil
	}

	s.running = false

	if s.delayTimer != nil {
		s.delayTimer.Stop()
	}
	if s.delayTimeoutTimer != nil {
		s.delayTimeoutTimer.Stop()
	}

	if s.closeFn != nil {
		err := s.closeFn()
		logrus.Info("串口已关闭: ", s.config.Port)
		return err
	}

	return nil
}

// readLoop 读取循环（模拟实现）
func (s *SerialPort) readLoop() {
	buf := make([]byte, 4096)

	for s.running {
		n := 0
		time.Sleep(10 * time.Millisecond)

		if n > 0 {
			data := buf[:n]
			s.handleData(data)
		}
	}
}

// handleData 处理接收到的数据，支持延迟封包
func (s *SerialPort) handleData(data []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.config.DelayPackaging > 0 {
		s.buffer.Write(data)

		if s.delayTimer != nil {
			s.delayTimer.Stop()
		}
		s.delayTimer = time.AfterFunc(time.Duration(s.config.DelayPackaging)*time.Millisecond, func() {
			s.flushBuffer()
		})

		if s.delayTimeoutTimer != nil {
			s.delayTimeoutTimer.Stop()
		}
		s.delayTimeoutTimer = time.AfterFunc(time.Duration(s.config.DelayTimeout)*time.Millisecond, func() {
			s.flushBuffer()
		})
	} else {
		if s.onData != nil {
			s.onData(data)
		}
	}
}

// flushBuffer 刷新缓冲区，发送缓存的数据
func (s *SerialPort) flushBuffer() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.buffer.Len() > 0 {
		data := make([]byte, s.buffer.Len())
		s.buffer.Read(data)

		if s.onData != nil {
			s.onData(data)
		}

		logrus.Debug("串口数据: ", hex.EncodeToString(data))
	}
}

// Write 向串口写入数据
func (s *SerialPort) Write(data []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return ErrNotOpen
	}

	logrus.Debug("串口写入: ", hex.EncodeToString(data))

	return nil
}

// SetOnData 设置数据接收回调
func (s *SerialPort) SetOnData(callback func([]byte)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.onData = callback
}

// IsOpen 检查串口是否打开
func (s *SerialPort) IsOpen() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.running
}

// GetConfig 获取串口配置
func (s *SerialPort) GetConfig() *SerialConfig {
	return s.config
}

// ErrNotOpen 串口未打开错误
var ErrNotOpen = &Error{Code: "not_open", Message: "串口未打开"}

// Error 串口错误类型
type Error struct {
	Code    string // 错误代码
	Message string // 错误消息
}

// Error 返回错误消息
func (e *Error) Error() string {
	return e.Message
}
