package system

import (
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sirupsen/logrus"
)

// SystemManager 系统管理器
type SystemManager struct {
	mutex        sync.Mutex    // 互斥锁
	shutdownChan chan struct{} // 关闭信号通道
	restartChan  chan struct{} // 重启信号通道
	upgradeChan  chan string   // 升级文件路径通道
}

// NewSystemManager 创建系统管理器实例
func NewSystemManager() *SystemManager {
	return &SystemManager{
		shutdownChan: make(chan struct{}),
		restartChan:  make(chan struct{}),
		upgradeChan:  make(chan string),
	}
}

// Start 启动系统管理器协程
func (s *SystemManager) Start() {
	go func() {
		for {
			select {
			case <-s.restartChan:
				s.Restart()
			case path := <-s.upgradeChan:
				s.Upgrade(path)
			case <-s.shutdownChan:
				return
			}
		}
	}()
}

// Shutdown 关闭系统管理器
func (s *SystemManager) Shutdown() {
	close(s.shutdownChan)
}

// Restart 重启系统
func (s *SystemManager) Restart() {
	logrus.Info("系统正在重启...")

	time.Sleep(1 * time.Second)

	if runtime.GOOS == "windows" {
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.Start()
		os.Exit(0)
	} else {
		args := strings.Join(os.Args[1:], " ")
		cmd := exec.Command("/bin/sh", "-c", "sleep 1 && exec "+os.Args[0]+" "+args)
		cmd.Start()
		os.Exit(0)
	}
}

// RequestRestart 请求重启系统
func (s *SystemManager) RequestRestart() {
	s.restartChan <- struct{}{}
}

// Upgrade 系统升级
func (s *SystemManager) Upgrade(filePath string) {
	logrus.Info("系统正在从 ", filePath, " 升级...")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logrus.Error("升级文件不存在: ", filePath)
		return
	}

	executablePath, err := os.Executable()
	if err != nil {
		logrus.Error("获取可执行文件路径失败: ", err)
		return
	}

	backupPath := executablePath + ".bak"
	if err := os.Rename(executablePath, backupPath); err != nil {
		logrus.Error("备份可执行文件失败: ", err)
		return
	}

	if err := copyFile(filePath, executablePath); err != nil {
		logrus.Error("复制升级文件失败: ", err)
		os.Rename(backupPath, executablePath)
		return
	}

	if err := os.Chmod(executablePath, 0755); err != nil {
		logrus.Error("设置可执行文件权限失败: ", err)
		return
	}

	logrus.Info("升级完成，正在重启...")
	s.RequestRestart()
}

// RequestUpgrade 请求系统升级
func (s *SystemManager) RequestUpgrade(filePath string) {
	s.upgradeChan <- filePath
}

// Reset 系统重置（恢复出厂设置）
func (s *SystemManager) Reset() error {
	configPath := "config.yaml"
	if _, err := os.Stat(configPath); err == nil {
		if err := os.Remove(configPath); err != nil {
			logrus.Error("删除配置文件失败: ", err)
			return err
		}
	}

	cachePath := "data/cache.dat"
	if _, err := os.Stat(cachePath); err == nil {
		if err := os.Remove(cachePath); err != nil {
			logrus.Error("删除缓存文件失败: ", err)
			return err
		}
	}

	logrus.Info("系统重置完成")
	return nil
}

// GetSystemInfo 获取系统信息
func (s *SystemManager) GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}

	hostname, err := os.Hostname()
	if err != nil {
		logrus.Error("获取主机名失败: ", err)
		return nil, err
	}
	info.Hostname = hostname

	hostInfo, err := host.Info()
	if err != nil {
		logrus.Error("获取主机信息失败: ", err)
		return nil, err
	}
	info.OS = hostInfo.OS
	info.Platform = hostInfo.Platform
	info.PlatformVersion = hostInfo.PlatformVersion
	info.KernelVersion = hostInfo.KernelVersion
	info.Uptime = hostInfo.Uptime

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		logrus.Error("获取内存信息失败: ", err)
		return nil, err
	}
	info.MemTotal = memInfo.Total
	info.MemUsed = memInfo.Used
	info.MemFree = memInfo.Free

	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		logrus.Error("获取CPU使用率失败: ", err)
		return nil, err
	}
	if len(cpuPercent) > 0 {
		info.CPUPercent = cpuPercent[0]
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		logrus.Error("获取CPU信息失败: ", err)
		return nil, err
	}
	if len(cpuInfo) > 0 {
		info.CPUModel = cpuInfo[0].ModelName
		info.CPUCores = int(cpuInfo[0].Cores)
	}

	diskInfo, err := disk.Usage("/")
	if err != nil {
		logrus.Error("获取磁盘信息失败: ", err)
		return nil, err
	}
	info.DiskTotal = diskInfo.Total
	info.DiskUsed = diskInfo.Used
	info.DiskFree = diskInfo.Free

	info.Architecture = runtime.GOARCH
	info.GoVersion = runtime.Version()

	return info, nil
}

// GetNetworkInfo 获取网络接口信息
func (s *SystemManager) GetNetworkInfo() ([]NetworkInterface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		logrus.Error("获取网络接口失败: ", err)
		return nil, err
	}

	var result []NetworkInterface
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var ipAddresses []string
		for _, addr := range addrs {
			ipAddresses = append(ipAddresses, addr.String())
		}

		result = append(result, NetworkInterface{
			Name:         iface.Name,
			HardwareAddr: iface.HardwareAddr.String(),
			IPAddresses:  ipAddresses,
			MTU:          iface.MTU,
			Flags:        iface.Flags.String(),
		})
	}

	return result, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, 64*1024)
	for {
		n, err := srcFile.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := dstFile.Write(buf[:n]); err != nil {
			return err
		}
	}

	return nil
}

// SystemInfo 系统信息结构体
type SystemInfo struct {
	Hostname        string  `json:"hostname"`         // 主机名
	OS              string  `json:"os"`               // 操作系统
	Platform        string  `json:"platform"`         // 平台
	PlatformVersion string  `json:"platform_version"` // 平台版本
	KernelVersion   string  `json:"kernel_version"`   // 内核版本
	Uptime          uint64  `json:"uptime"`           // 运行时间(秒)
	MemTotal        uint64  `json:"mem_total"`        // 总内存
	MemUsed         uint64  `json:"mem_used"`         // 已用内存
	MemFree         uint64  `json:"mem_free"`         // 空闲内存
	CPUPercent      float64 `json:"cpu_percent"`      // CPU使用率
	CPUModel        string  `json:"cpu_model"`        // CPU型号
	CPUCores        int     `json:"cpu_cores"`        // CPU核心数
	DiskTotal       uint64  `json:"disk_total"`       // 总磁盘空间
	DiskUsed        uint64  `json:"disk_used"`        // 已用磁盘空间
	DiskFree        uint64  `json:"disk_free"`        // 空闲磁盘空间
	Architecture    string  `json:"architecture"`     // 架构
	GoVersion       string  `json:"go_version"`       // Go版本
}

// NetworkInterface 网络接口信息结构体
type NetworkInterface struct {
	Name         string   `json:"name"`          // 接口名称
	HardwareAddr string   `json:"hardware_addr"` // MAC地址
	IPAddresses  []string `json:"ip_addresses"`  // IP地址列表
	MTU          int      `json:"mtu"`           // MTU
	Flags        string   `json:"flags"`         // 标志位
}
