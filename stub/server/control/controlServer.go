package control

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gns/stub/env"
	"gns/stub/log"
	"gns/stub/server/managed"
	"gns/stub/server/managed/entities"
	"gns/stub/server/managed/servers"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

// ControlServer TODO Latency
type ControlServer struct {
	ManagedServer managed.ManagedServerInterface
	ControlServer *http.Server
	mu            sync.RWMutex
	addr          string
	port          string
	log           *zap.Logger
}

func NewControlServer(env env.Environment) *ControlServer {
	cs := &ControlServer{
		addr: env.ControlServerAddr,
		port: env.ControlServerPort,
		log:  log.InitLogger(env.LogLevel),
	}
	switch env.ManagedServerType {
	//case managed.ServerTypeFastHTTP:
	//	cs.ManagedServer = servers.NewFastHTTPServer(env)
	//case managed.ServerTypeGin:
	//	cs.ManagedServer = servers.NewGinServer(env)
	default:
		cs.ManagedServer = servers.NewNetHttpServer(env)
	}

	cs.log.Debug("Managed server type", zap.String("type", env.ManagedServerType))
	return cs
}

func (s *ControlServer) InitControlServer() {
	gin.SetMode(gin.ReleaseMode)
	s.log.Debug("Initializing control server", zap.String("address", s.addr), zap.String("port", s.port))

	// Инициализация роутеров и эндпоинтов
	r := gin.New()
	r.GET("/rest/api/v1/server/config", s.GetControlServerConfig)
	r.POST("/rest/api/v1/server/config", s.UpdateControlServerConfig)
	r.DELETE("/rest/api/v1/server/config", s.DeleteControlServerConfig)
	r.POST("/rest/api/v1/server/start", s.StartManagedServer)
	r.POST("/rest/api/v1/server/stop", s.StopManagedServer)
	r.GET("/rest/api/v1/server/status", s.StatusControlServer)

	s.ControlServer = &http.Server{
		Addr:           fmt.Sprintf("%s:%s", s.addr, s.port),
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func (s *ControlServer) RunControlServer() {
	s.log.Info("Running control server",
		zap.String("address", s.addr),
		zap.String("port", s.port))
	if err := s.ControlServer.ListenAndServe(); err != nil {
		s.log.Fatal("Failed to start control server", zap.Error(err))
	}
}

func (s *ControlServer) GetControlServerConfig(c *gin.Context) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	config, exist := s.ManagedServer.GetConfig()
	if !exist {
		c.JSON(http.StatusOK, nil)
		return
	}
	c.JSON(http.StatusOK, config)

}

func (s *ControlServer) UpdateControlServerConfig(c *gin.Context) {
	var newConfig entities.ServerConfig
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration"})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.ManagedServer.SetConfig(newConfig)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.String(http.StatusOK, "The configuration has been successfully applied")
	s.log.Debug("The configuration of managed server has been successfully applied")
}

func (s *ControlServer) DeleteControlServerConfig(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.ManagedServer.SetConfig(entities.ServerConfig{})
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		s.log.Error("Failed to delete configuration", zap.Error(err))
		return
	}
	c.String(http.StatusOK, "Configuration deleted successfully")
	s.log.Debug("Managed server config deleted")
}

func (s *ControlServer) StartManagedServer(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ManagedServer.IsRunning() {
		s.log.Debug("Attempted to start managed server as already running")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Server already running"})
		return
	}

	s.ManagedServer.SetRunning(true)
	c.String(http.StatusOK, "Server started successfully")
	s.log.Debug("Managed server started successfully")
}

func (s *ControlServer) StopManagedServer(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.ManagedServer.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Server is not running"})
		return
	}

	s.ManagedServer.SetRunning(false)
	c.String(http.StatusOK, "Server stopped successfully")
	s.log.Debug("Managed server stopped successfully")
}

func (s *ControlServer) StatusControlServer(c *gin.Context) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	seconds := time.Since(s.ManagedServer.GetTimeSinceStart()).Seconds()

	status := struct {
		Running    bool    `json:"running"`
		TPS        uint    `json:"tps"`
		AvgLatency float64 `json:"avg_latency"`
		Duration   float64 `json:"duration"`
	}{
		Running:    s.ManagedServer.IsRunning(),
		TPS:        s.ManagedServer.GetReqSinceStart() / uint(seconds),
		AvgLatency: 1.0, // Заглушка для средней задержки
		Duration:   seconds,
	}
	c.JSON(http.StatusOK, status)
	s.log.Debug("Managed server status", zap.Any("status", status))
}

func (s *ControlServer) InitManagedServer() {
	s.ManagedServer.InitManagedServer() // Инициализация сервера
}

func (s *ControlServer) RunManagedServer() {
	s.ManagedServer.RunManagedServer() // Запуск сервера
}
