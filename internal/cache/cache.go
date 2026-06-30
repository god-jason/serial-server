package cache

import (
	"encoding/binary"
	"os"
	"sync"
	"time"

	"github.com/golang/snappy"
	"github.com/sirupsen/logrus"
)

// DataCache 数据缓存管理器
type DataCache struct {
	mutex   sync.RWMutex     // 读写锁
	file    *os.File        // 缓存文件句柄
	queue   [][]byte        // 缓存队列
	maxSize int             // 最大缓存数量
	maxAge  time.Duration   // 最大缓存时间
}

// NewDataCache 创建数据缓存实例
func NewDataCache(filePath string, maxSize int, maxAge time.Duration) (*DataCache, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		logrus.Error("打开缓存文件失败: ", err)
		return nil, err
	}

	cache := &DataCache{
		file:    file,
		maxSize: maxSize,
		maxAge:  maxAge,
		queue:   make([][]byte, 0),
	}

	cache.loadFromFile()

	return cache, nil
}

// loadFromFile 从文件加载缓存数据
func (c *DataCache) loadFromFile() {
	data, err := os.ReadFile(c.file.Name())
	if err != nil {
		return
	}

	if len(data) == 0 {
		return
	}

	decoded, err := snappy.Decode(nil, data)
	if err != nil {
		logrus.Error("解码缓存文件失败: ", err)
		return
	}

	offset := 0
	for offset < len(decoded) {
		if offset+8 > len(decoded) {
			break
		}

		timestamp := int64(binary.LittleEndian.Uint64(decoded[offset : offset+8]))
		offset += 8

		if offset+4 > len(decoded) {
			break
		}

		length := int(binary.LittleEndian.Uint32(decoded[offset : offset+4]))
		offset += 4

		if offset+length > len(decoded) {
			break
		}

		dataItem := make([]byte, length)
		copy(dataItem, decoded[offset:offset+length])
		offset += length

		if time.Since(time.Unix(0, timestamp)) < c.maxAge {
			c.queue = append(c.queue, dataItem)
		}
	}

	logrus.Info("已加载 ", len(c.queue), " 条缓存数据")
}

// Add 添加数据到缓存
func (c *DataCache) Add(data []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.queue) >= c.maxSize {
		c.queue = c.queue[1:]
	}

	c.queue = append(c.queue, data)

	return c.saveToFile()
}

// saveToFile 将缓存数据保存到文件
func (c *DataCache) saveToFile() error {
	var buf []byte
	for _, item := range c.queue {
		timestamp := time.Now().UnixNano()
		header := make([]byte, 12)
		binary.LittleEndian.PutUint64(header[0:8], uint64(timestamp))
		binary.LittleEndian.PutUint32(header[8:12], uint32(len(item)))

		buf = append(buf, header...)
		buf = append(buf, item...)
	}

	compressed := snappy.Encode(nil, buf)

	if err := os.Truncate(c.file.Name(), 0); err != nil {
		logrus.Error("截断缓存文件失败: ", err)
		return err
	}

	if _, err := c.file.Seek(0, 0); err != nil {
		logrus.Error("文件定位失败: ", err)
		return err
	}

	_, err := c.file.Write(compressed)
	if err != nil {
		logrus.Error("写入缓存文件失败: ", err)
	}
	return err
}

// GetAll 获取所有缓存数据
func (c *DataCache) GetAll() [][]byte {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	result := make([][]byte, len(c.queue))
	copy(result, c.queue)
	return result
}

// Remove 移除指定索引的缓存数据
func (c *DataCache) Remove(index int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if index >= 0 && index < len(c.queue) {
		c.queue = append(c.queue[:index], c.queue[index+1:]...)
		c.saveToFile()
	}
}

// RemoveAll 清空所有缓存数据
func (c *DataCache) RemoveAll() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.queue = make([][]byte, 0)
	c.saveToFile()
}

// Count 获取缓存数据数量
func (c *DataCache) Count() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.queue)
}

// Close 关闭缓存文件
func (c *DataCache) Close() error {
	return c.file.Close()
}
