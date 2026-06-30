package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/god-jason/serial-server/internal/config"
	"github.com/god-jason/serial-server/internal/server"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const version = "1.0.0"

// main 程序入口函数
func main() {
	setupLogging()

	logrus.Info("串口服务器 v", version, " 正在启动...")

	if err := loadConfig(); err != nil {
		logrus.Fatal("加载配置失败: ", err)
	}

	srv := server.NewServer()

	go func() {
		if err := srv.Start(); err != nil && err != server.ErrServerClosed {
			logrus.Fatal("服务器启动失败: ", err)
		}
	}()

	config.Watch(func() {
		logrus.Info("配置文件已重载，正在重启服务器...")
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logrus.Info("正在关闭服务器...")

	if err := srv.Stop(); err != nil {
		logrus.Error("服务器关闭失败: ", err)
	}

	logrus.Info("服务器已停止")
}

// setupLogging 配置日志系统
func setupLogging() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	os.MkdirAll("logs", 0755)
	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logrus.Error("打开日志文件失败: ", err)
		return
	}

	logrus.SetOutput(file)
}

// loadConfig 加载配置文件
func loadConfig() error {
	if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
		cfg := config.GetDefault()
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		if err := os.WriteFile("config.yaml", data, 0644); err != nil {
			return err
		}
		logrus.Info("已创建默认配置文件 config.yaml")
	}

	return config.Load("config.yaml")
}
