package modbus

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// ModbusConverter Modbus协议转换器
type ModbusConverter struct {
	mutex        sync.Mutex       // 互斥锁
	serialOnData func([]byte)     // 串口数据发送回调
	tcpOnData    func([]byte)     // TCP数据发送回调
	buffer       bytes.Buffer     // 数据缓冲区
}

// NewModbusConverter 创建Modbus协议转换器实例
func NewModbusConverter() *ModbusConverter {
	return &ModbusConverter{}
}

// SetSerialOnData 设置串口数据发送回调
func (m *ModbusConverter) SetSerialOnData(callback func([]byte)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.serialOnData = callback
}

// SetTCPOnData 设置TCP数据发送回调
func (m *ModbusConverter) SetTCPOnData(callback func([]byte)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.tcpOnData = callback
}

// HandleSerialData 处理串口数据，将RTU转换为TCP
func (m *ModbusConverter) HandleSerialData(data []byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	rtuFrame, err := parseRTUFrame(data)
	if err != nil {
		logrus.Debug("解析RTU帧失败: ", err)
		return
	}

	tcpFrame := rtuToTCP(rtuFrame)

	logrus.Debug("Modbus RTU转换为TCP: ", hexEncode(tcpFrame))

	if m.tcpOnData != nil {
		m.tcpOnData(tcpFrame)
	}
}

// HandleTCPData 处理TCP数据，将TCP转换为RTU
func (m *ModbusConverter) HandleTCPData(data []byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	tcpFrame, err := parseTCPFrame(data)
	if err != nil {
		logrus.Debug("解析TCP帧失败: ", err)
		return
	}

	rtuFrame := tcpToRTU(tcpFrame)

	logrus.Debug("Modbus TCP转换为RTU: ", hexEncode(rtuFrame))

	if m.serialOnData != nil {
		m.serialOnData(rtuFrame)
	}
}

// parseRTUFrame 解析Modbus RTU帧
func parseRTUFrame(data []byte) (*RTUFrame, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("帧长度过短")
	}

	frame := &RTUFrame{
		SlaveAddress: data[0],
		FunctionCode: data[1],
	}

	switch frame.FunctionCode {
	case 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x0B, 0x0C:
		if len(data) >= 5 {
			frame.Data = data[2 : len(data)-2]
			frame.CRC = binary.LittleEndian.Uint16(data[len(data)-2:])
		}
	case 0x0F, 0x10:
		if len(data) >= 7 {
			frame.Data = data[2 : len(data)-2]
			frame.CRC = binary.LittleEndian.Uint16(data[len(data)-2:])
		}
	default:
		if len(data) >= 3 {
			frame.Data = data[2 : len(data)-2]
			frame.CRC = binary.LittleEndian.Uint16(data[len(data)-2:])
		}
	}

	computedCRC := crc16(data[:len(data)-2])
	if frame.CRC != computedCRC {
		return nil, fmt.Errorf("CRC校验失败")
	}

	return frame, nil
}

// parseTCPFrame 解析Modbus TCP帧
func parseTCPFrame(data []byte) (*TCPFrame, error) {
	if len(data) < 7 {
		return nil, fmt.Errorf("帧长度过短")
	}

	frame := &TCPFrame{
		TransactionID: binary.BigEndian.Uint16(data[0:2]),
		ProtocolID:    binary.BigEndian.Uint16(data[2:4]),
		Length:        binary.BigEndian.Uint16(data[4:6]),
		UnitID:        data[6],
	}

	if int(frame.Length)+6 != len(data) {
		return nil, fmt.Errorf("长度不匹配")
	}

	if len(data) > 7 {
		frame.Data = data[7:]
	}

	return frame, nil
}

// rtuToTCP 将RTU帧转换为TCP帧
func rtuToTCP(rtu *RTUFrame) []byte {
	length := uint16(len(rtu.Data) + 1)

	buf := make([]byte, 7+len(rtu.Data))
	binary.BigEndian.PutUint16(buf[0:2], 0)
	binary.BigEndian.PutUint16(buf[2:4], 0)
	binary.BigEndian.PutUint16(buf[4:6], length)
	buf[6] = rtu.SlaveAddress
	copy(buf[7:], rtu.Data)

	return buf
}

// tcpToRTU 将TCP帧转换为RTU帧
func tcpToRTU(tcp *TCPFrame) []byte {
	buf := make([]byte, 2+len(tcp.Data)+2)
	buf[0] = tcp.UnitID
	copy(buf[1:], tcp.Data)

	crc := crc16(buf[:len(buf)-2])
	binary.LittleEndian.PutUint16(buf[len(buf)-2:], crc)

	return buf
}

// crc16 计算CRC16校验值
func crc16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

// hexEncode 将字节数组转换为十六进制字符串
func hexEncode(data []byte) string {
	result := ""
	for i, b := range data {
		if i > 0 {
			result += " "
		}
		result += fmt.Sprintf("%02X", b)
	}
	return result
}

// RTUFrame Modbus RTU帧结构
type RTUFrame struct {
	SlaveAddress byte   // 从站地址
	FunctionCode byte   // 功能码
	Data         []byte // 数据
	CRC          uint16 // CRC校验
}

// TCPFrame Modbus TCP帧结构
type TCPFrame struct {
	TransactionID uint16 // 事务ID
	ProtocolID    uint16 // 协议ID
	Length        uint16 // 长度
	UnitID        byte   // 单元ID
	Data          []byte // 数据
}
