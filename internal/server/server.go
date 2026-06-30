package server

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/god-jason/serial-server/internal/config"
	"github.com/god-jason/serial-server/internal/serial"
	"github.com/god-jason/serial-server/internal/system"
	"github.com/god-jason/serial-server/internal/websocket"
	gorilla "github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// ErrServerClosed 服务器关闭错误
var ErrServerClosed = fmt.Errorf("服务器已关闭")

// Server HTTP服务器结构体
type Server struct {
	gin         *gin.Engine                   // Gin引擎
	httpServer  *http.Server                  // HTTP服务器
	wsManager   *websocket.WebSocketManager   // WebSocket管理器
	systemMgr   *system.SystemManager         // 系统管理器
	serialPorts map[string]*serial.SerialPort // 串口端口映射
	mutex       sync.Mutex                    // 互斥锁
	jwtSecret   string                        // JWT密钥
	stats       Stats                         // 统计信息
}

// Stats 流量统计结构体
type Stats struct {
	SerialTX          uint64 `json:"serial_tx"`          // 串口发送字节数
	SerialRX          uint64 `json:"serial_rx"`          // 串口接收字节数
	NetworkTX         uint64 `json:"network_tx"`         // 网络发送字节数
	NetworkRX         uint64 `json:"network_rx"`         // 网络接收字节数
	CacheCount        int    `json:"cache_count"`        // 缓存数据数量
	ResendCount       int    `json:"resend_count"`       // 重发数据数量
	ConnectedChannels int    `json:"connected_channels"` // 已连接通道数
	OpenedPorts       int    `json:"opened_ports"`       // 已打开串口号
}

// NewServer 创建服务器实例
func NewServer() *Server {
	return &Server{
		wsManager:   websocket.NewWebSocketManager(),
		systemMgr:   system.NewSystemManager(),
		serialPorts: make(map[string]*serial.SerialPort),
		jwtSecret:   "serial-server-secret",
	}
}

// Start 启动HTTP服务器
func (s *Server) Start() error {
	s.gin = gin.Default()

	s.setupMiddleware()
	s.setupRoutes()

	cfg := config.Get()
	addr := fmt.Sprintf(":%d", cfg.System.WebPort)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.gin,
	}

	logrus.Info("HTTP服务器正在端口 ", cfg.System.WebPort, " 启动...")

	s.systemMgr.Start()

	return s.httpServer.ListenAndServe()
}

// Stop 停止HTTP服务器
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.systemMgr.Shutdown()

	return s.httpServer.Shutdown(ctx)
}

// setupMiddleware 设置中间件
func (s *Server) setupMiddleware() {
	s.gin.Use(gin.Logger())
	s.gin.Use(gin.Recovery())

	store := cookie.NewStore([]byte(s.jwtSecret))
	store.Options(sessions.Options{
		MaxAge: 86400,
		Path:   "/",
	})
	s.gin.Use(sessions.Sessions("serial-server-session", store))
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	s.gin.POST("/api/login", s.login)
	s.gin.PUT("/api/password", s.changePassword)

	api := s.gin.Group("/api")
	api.Use(s.authMiddleware())

	api.GET("/config", s.getConfig)
	api.PUT("/config", s.updateConfig)

	api.GET("/system/info", s.getSystemInfo)
	api.GET("/system/network", s.getNetworkInfo)
	api.POST("/system/restart", s.restartSystem)
	api.POST("/system/reset", s.resetSystem)
	api.POST("/system/upgrade", s.upgradeSystem)

	api.GET("/serial/ports", s.listSerialPorts)
	api.POST("/serial/ports", s.addSerialPort)
	api.PUT("/serial/ports/:id", s.updateSerialPort)
	api.DELETE("/serial/ports/:id", s.deleteSerialPort)
	api.POST("/serial/ports/:id/open", s.openSerialPort)
	api.POST("/serial/ports/:id/close", s.closeSerialPort)

	api.GET("/channels", s.listChannels)
	api.POST("/channels", s.addChannel)
	api.PUT("/channels/:id", s.updateChannel)
	api.DELETE("/channels/:id", s.deleteChannel)
	api.POST("/channels/:id/enable", s.enableChannel)
	api.POST("/channels/:id/disable", s.disableChannel)

	api.GET("/channels/tcp-client", s.listTCPClientChannels)
	api.POST("/channels/tcp-client", s.addTCPClientChannel)
	api.PUT("/channels/tcp-client/:id", s.updateTCPClientChannel)
	api.DELETE("/channels/tcp-client/:id", s.deleteTCPClientChannel)

	api.GET("/channels/tcp-server", s.listTCPServerChannels)
	api.POST("/channels/tcp-server", s.addTCPServerChannel)
	api.PUT("/channels/tcp-server/:id", s.updateTCPServerChannel)
	api.DELETE("/channels/tcp-server/:id", s.deleteTCPServerChannel)

	api.GET("/channels/mqtt", s.listMQTTChannels)
	api.POST("/channels/mqtt", s.addMQTTChannel)
	api.PUT("/channels/mqtt/:id", s.updateMQTTChannel)
	api.DELETE("/channels/mqtt/:id", s.deleteMQTTChannel)

	api.GET("/logs", s.getLogs)
	api.GET("/stats", s.getStats)

	s.gin.GET("/ws/serial/:port", s.websocketSerial)
	s.gin.GET("/ws/terminal", s.websocketTerminal)

	s.gin.NoRoute(func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path, http.FS(frontendSubFS))
	})
}

// WebSocket升级器配置
var upgrader = gorilla.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// authMiddleware 认证中间件
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		login := session.Get("login")
		if login != true {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			return
		}
		c.Next()
	}
}

// login 登录接口
func (s *Server) login(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	cfg := config.Get()
	if req.Password == cfg.System.Password {
		session := sessions.Default(c)
		session.Set("login", true)
		session.Set("ip", c.ClientIP())
		session.Save()
		c.JSON(http.StatusOK, gin.H{"success": true})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
	}
}

// changePassword 修改密码接口
func (s *Server) changePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	cfg := config.Get()
	if req.OldPassword != cfg.System.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "旧密码错误"})
		return
	}

	cfg.System.Password = req.NewPassword
	config.Save()

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// getConfig 获取配置接口
func (s *Server) getConfig(c *gin.Context) {
	cfg := config.Get()
	c.JSON(http.StatusOK, cfg)
}

// updateConfig 更新配置接口
func (s *Server) updateConfig(c *gin.Context) {
	var newConfig config.Config
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	config.Update(&newConfig)
	config.Save()

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// getSystemInfo 获取系统信息接口
func (s *Server) getSystemInfo(c *gin.Context) {
	info, err := s.systemMgr.GetSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

// getNetworkInfo 获取网络信息接口
func (s *Server) getNetworkInfo(c *gin.Context) {
	info, err := s.systemMgr.GetNetworkInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

// restartSystem 重启系统接口
func (s *Server) restartSystem(c *gin.Context) {
	go s.systemMgr.RequestRestart()
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "系统正在重启"})
}

// resetSystem 重置系统接口
func (s *Server) resetSystem(c *gin.Context) {
	if err := s.systemMgr.Reset(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	go s.systemMgr.RequestRestart()
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "系统已重置，正在重启..."})
}

// upgradeSystem 系统升级接口
func (s *Server) upgradeSystem(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未上传文件"})
		return
	}
	defer file.Close()

	tmpFile, err := os.CreateTemp("", "upgrade-*.bin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go s.systemMgr.RequestUpgrade(tmpFile.Name())
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "升级已开始"})
}

// listSerialPorts 列出串口接口
func (s *Server) listSerialPorts(c *gin.Context) {
	cfg := config.Get()
	c.JSON(http.StatusOK, cfg.Serial.Ports)
}

// addSerialPort 添加串口接口
func (s *Server) addSerialPort(c *gin.Context) {
	var port config.SerialPortConfig
	if err := c.ShouldBindJSON(&port); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	cfg := config.Get()
	cfg.Serial.Ports = append(cfg.Serial.Ports, port)
	config.Save()

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// updateSerialPort 更新串口接口
func (s *Server) updateSerialPort(c *gin.Context) {
	id := c.Param("id")
	var port config.SerialPortConfig
	if err := c.ShouldBindJSON(&port); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	cfg := config.Get()
	for i, p := range cfg.Serial.Ports {
		if p.ID == id {
			cfg.Serial.Ports[i] = port
			config.Save()
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "串口未找到"})
}

// deleteSerialPort 删除串口接口
func (s *Server) deleteSerialPort(c *gin.Context) {
	id := c.Param("id")

	cfg := config.Get()
	for i, p := range cfg.Serial.Ports {
		if p.ID == id {
			cfg.Serial.Ports = append(cfg.Serial.Ports[:i], cfg.Serial.Ports[i+1:]...)
			config.Save()
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "串口未找到"})
}

// openSerialPort 打开串口接口
func (s *Server) openSerialPort(c *gin.Context) {
	id := c.Param("id")

	cfg := config.Get()
	for _, p := range cfg.Serial.Ports {
		if p.ID == id {
			serialConfig := &serial.SerialConfig{
				ID:             p.ID,
				Name:           p.Name,
				Port:           p.Port,
				BaudRate:       p.BaudRate,
				DataBits:       p.DataBits,
				Parity:         p.Parity,
				StopBits:       p.StopBits,
				FlowControl:    p.FlowControl,
				DelayPackaging: p.DelayPackaging,
				DelayTimeout:   p.DelayTimeout,
				Protocol:       p.Protocol,
				Enabled:        p.Enabled,
			}

			sp := serial.NewSerialPort(serialConfig)
			if err := sp.Open(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			s.mutex.Lock()
			s.serialPorts[id] = sp
			s.stats.OpenedPorts = len(s.serialPorts)
			s.mutex.Unlock()

			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "串口未找到"})
}

// closeSerialPort 关闭串口接口
func (s *Server) closeSerialPort(c *gin.Context) {
	id := c.Param("id")

	s.mutex.Lock()
	sp, ok := s.serialPorts[id]
	s.mutex.Unlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "串口未找到"})
		return
	}

	if err := sp.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.mutex.Lock()
	delete(s.serialPorts, id)
	s.stats.OpenedPorts = len(s.serialPorts)
	s.mutex.Unlock()

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// listChannels 列出通道接口
func (s *Server) listChannels(c *gin.Context) {
	cfg := config.Get()
	c.JSON(http.StatusOK, cfg.Channels)
}

// addChannel 添加通道接口
func (s *Server) addChannel(c *gin.Context) {
	var channel config.ChannelConfig
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	cfg := config.Get()
	cfg.Channels = append(cfg.Channels, channel)
	config.Save()

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// updateChannel 更新通道接口
func (s *Server) updateChannel(c *gin.Context) {
	id := c.Param("id")
	var channel config.ChannelConfig
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	cfg := config.Get()
	for i, ch := range cfg.Channels {
		if ch.ID == id {
			cfg.Channels[i] = channel
			config.Save()
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "通道未找到"})
}

// deleteChannel 删除通道接口
func (s *Server) deleteChannel(c *gin.Context) {
	id := c.Param("id")

	cfg := config.Get()
	for i, ch := range cfg.Channels {
		if ch.ID == id {
			cfg.Channels = append(cfg.Channels[:i], cfg.Channels[i+1:]...)
			config.Save()
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "通道未找到"})
}

// enableChannel 启用通道接口
func (s *Server) enableChannel(c *gin.Context) {
	id := c.Param("id")

	cfg := config.Get()
	for i, ch := range cfg.Channels {
		if ch.ID == id {
			cfg.Channels[i].Enabled = true
			config.Save()
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "通道未找到"})
}

// disableChannel 禁用通道接口
func (s *Server) disableChannel(c *gin.Context) {
	id := c.Param("id")

	cfg := config.Get()
	for i, ch := range cfg.Channels {
		if ch.ID == id {
			cfg.Channels[i].Enabled = false
			config.Save()
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "通道未找到"})
}

// listTCPClientChannels 列出TCP客户端通道
func (s *Server) listTCPClientChannels(c *gin.Context) {
	cfg := config.Get()
	var tcpClientChannels []config.ChannelConfig
	for _, ch := range cfg.Channels {
		if ch.Type == "tcp_client" {
			tcpClientChannels = append(tcpClientChannels, ch)
		}
	}
	c.JSON(http.StatusOK, tcpClientChannels)
}

// addTCPClientChannel 添加TCP客户端通道
func (s *Server) addTCPClientChannel(c *gin.Context) {
	var channel config.ChannelConfig
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	channel.Type = "tcp_client"

	cfg := config.Get()
	cfg.Channels = append(cfg.Channels, channel)
	config.Save()

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// updateTCPClientChannel 更新TCP客户端通道
func (s *Server) updateTCPClientChannel(c *gin.Context) {
	id := c.Param("id")
	var channel config.ChannelConfig
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	channel.Type = "tcp_client"

	cfg := config.Get()
	for i, ch := range cfg.Channels {
		if ch.ID == id && ch.Type == "tcp_client" {
			cfg.Channels[i] = channel
			config.Save()
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "TCP客户端通道未找到"})
}

// deleteTCPClientChannel 删除TCP客户端通道
func (s *Server) deleteTCPClientChannel(c *gin.Context) {
	id := c.Param("id")

	cfg := config.Get()
	for i, ch := range cfg.Channels {
		if ch.ID == id && ch.Type == "tcp_client" {
			cfg.Channels = append(cfg.Channels[:i], cfg.Channels[i+1:]...)
			config.Save()
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "TCP客户端通道未找到"})
}

// listTCPServerChannels 列出TCP服务端通道
func (s *Server) listTCPServerChannels(c *gin.Context) {
	cfg := config.Get()
	var tcpServerChannels []config.ChannelConfig
	for _, ch := range cfg.Channels {
		if ch.Type == "tcp_server" {
			tcpServerChannels = append(tcpServerChannels, ch)
		}
	}
	c.JSON(http.StatusOK, tcpServerChannels)
}

// addTCPServerChannel 添加TCP服务端通道
func (s *Server) addTCPServerChannel(c *gin.Context) {
	var channel config.ChannelConfig
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	channel.Type = "tcp_server"

	cfg := config.Get()
	cfg.Channels = append(cfg.Channels, channel)
	config.Save()

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// updateTCPServerChannel 更新TCP服务端通道
func (s *Server) updateTCPServerChannel(c *gin.Context) {
	id := c.Param("id")
	var channel config.ChannelConfig
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	channel.Type = "tcp_server"

	cfg := config.Get()
	for i, ch := range cfg.Channels {
		if ch.ID == id && ch.Type == "tcp_server" {
			cfg.Channels[i] = channel
			config.Save()
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "TCP服务端通道未找到"})
}

// deleteTCPServerChannel 删除TCP服务端通道
func (s *Server) deleteTCPServerChannel(c *gin.Context) {
	id := c.Param("id")

	cfg := config.Get()
	for i, ch := range cfg.Channels {
		if ch.ID == id && ch.Type == "tcp_server" {
			cfg.Channels = append(cfg.Channels[:i], cfg.Channels[i+1:]...)
			config.Save()
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "TCP服务端通道未找到"})
}

// listMQTTChannels 列出MQTT通道
func (s *Server) listMQTTChannels(c *gin.Context) {
	cfg := config.Get()
	var mqttChannels []config.ChannelConfig
	for _, ch := range cfg.Channels {
		if ch.Type == "mqtt" {
			mqttChannels = append(mqttChannels, ch)
		}
	}
	c.JSON(http.StatusOK, mqttChannels)
}

// addMQTTChannel 添加MQTT通道
func (s *Server) addMQTTChannel(c *gin.Context) {
	var channel config.ChannelConfig
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	channel.Type = "mqtt"

	cfg := config.Get()
	cfg.Channels = append(cfg.Channels, channel)
	config.Save()

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// updateMQTTChannel 更新MQTT通道
func (s *Server) updateMQTTChannel(c *gin.Context) {
	id := c.Param("id")
	var channel config.ChannelConfig
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	channel.Type = "mqtt"

	cfg := config.Get()
	for i, ch := range cfg.Channels {
		if ch.ID == id && ch.Type == "mqtt" {
			cfg.Channels[i] = channel
			config.Save()
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "MQTT通道未找到"})
}

// deleteMQTTChannel 删除MQTT通道
func (s *Server) deleteMQTTChannel(c *gin.Context) {
	id := c.Param("id")

	cfg := config.Get()
	for i, ch := range cfg.Channels {
		if ch.ID == id && ch.Type == "mqtt" {
			cfg.Channels = append(cfg.Channels[:i], cfg.Channels[i+1:]...)
			config.Save()
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "MQTT通道未找到"})
}

// getLogs 获取日志接口
func (s *Server) getLogs(c *gin.Context) {
	cfg := config.Get()
	data, err := os.ReadFile(cfg.System.LogFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.String(http.StatusOK, string(data))
}

// getStats 获取统计信息接口
func (s *Server) getStats(c *gin.Context) {
	s.mutex.Lock()
	stats := s.stats
	s.mutex.Unlock()
	c.JSON(http.StatusOK, stats)
}

// websocketSerial WebSocket串口调试接口
func (s *Server) websocketSerial(c *gin.Context) {
	port := c.Param("port")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Error("WebSocket升级失败: ", err)
		return
	}

	s.wsManager.AddSerialConnection(port, conn)

	defer func() {
		s.wsManager.RemoveSerialConnection(port, conn)
		conn.Close()
	}()

	h := websocket.NewSerialDebugHandler(s.wsManager, nil)
	h.Handle(conn)
}

// websocketTerminal WebSocket终端接口
func (s *Server) websocketTerminal(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Error("WebSocket升级失败: ", err)
		return
	}

	s.wsManager.StartTerminal(conn)
}

//go:embed dist
var frontendFS embed.FS

var frontendSubFS = func() fs.FS {
	f, _ := fs.Sub(frontendFS, "dist")
	return f
}()
