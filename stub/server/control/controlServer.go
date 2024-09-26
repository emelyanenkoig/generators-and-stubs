package control

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gns/stub/env"
	"gns/stub/server/managed"
	"gns/stub/server/managed/entities"
	"gns/stub/server/managed/servers"
	"log"
	"net/http"
	"os"
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
}

func (cs *ControlServer) InitControlServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/rest/api/v1/server/config", cs.GetControlServerConfig)
	r.POST("/rest/api/v1/server/config", cs.UpdateControlServerConfig)

	r.DELETE("/rest/api/v1/server/config", cs.DeleteControlServerConfig)
	r.POST("/rest/api/v1/server/start", cs.StartManagedServer)
	r.POST("/rest/api/v1/server/stop", cs.StopManagedServer)
	r.GET("/rest/api/v1/server/status", cs.StatusControlServer)

	cs.ControlServer = &http.Server{
		Addr:           fmt.Sprintf("%s:%s", cs.addr, cs.port),
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func (cs *ControlServer) RunControlServer() {
	log.Printf("Starting control server on %s:%s", cs.addr, cs.port)
	if err := cs.ControlServer.ListenAndServe(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

}

func (cs *ControlServer) GetControlServerConfig(c *gin.Context) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	c.JSON(http.StatusOK, cs.ManagedServer.GetConfig())
}

func (cs *ControlServer) UpdateControlServerConfig(c *gin.Context) {
	var newConfig entities.ServerConfig
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration"})
		return
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.ManagedServer.SetConfig(newConfig)
	c.String(http.StatusOK, "The configuration has been successfully applied")
}

func (cs *ControlServer) DeleteControlServerConfig(c *gin.Context) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.ManagedServer.SetConfig(entities.ServerConfig{})
	c.String(http.StatusOK, "Configuration deleted successfully")
}

func (cs *ControlServer) StartManagedServer(c *gin.Context) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.ManagedServer.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Server already running"})
		return
	}

	cs.ManagedServer.SetRunning(true)
	go cs.ManagedServer.RunManagedServer() // запуск сервера в отдельной горутине
	c.String(http.StatusOK, "Server started successfully")
}

func (cs *ControlServer) StopManagedServer(c *gin.Context) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if !cs.ManagedServer.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Server is not running"})
		return
	}

	cs.ManagedServer.SetRunning(false)
	c.String(http.StatusOK, "Server stopped successfully")
}

func (cs *ControlServer) StatusControlServer(c *gin.Context) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	seconds := time.Since(cs.ManagedServer.GetTimeSinceStart()).Seconds()

	status := struct {
		Running    bool    `json:"running"`
		TPS        uint    `json:"tps"`
		AvgLatency float64 `json:"avg_latency"`
		Duration   float64 `json:"duration"`
	}{
		Running:    cs.ManagedServer.IsRunning(),
		TPS:        cs.ManagedServer.GetReqSinceStart() / uint(seconds),
		AvgLatency: 1.0, // Заглушка для средней задержки
		Duration:   seconds,
	}

	c.JSON(http.StatusOK, status)
}

func (cs *ControlServer) InitManagedServer() {
	cs.ManagedServer.InitManagedServer() // Инициализация сервера
}

func (cs *ControlServer) RunManagedServer() {
	cs.ManagedServer.RunManagedServer() // Запуск сервера
}

func NewControlServer(env env.Environment) *ControlServer {
	cs := &ControlServer{
		addr: env.ControlServerAddr,
		port: env.ControlServerPort,
	}
	switch env.ManagedServerType {
	case managed.ServerTypeFastHTTP:
		cs.ManagedServer = servers.NewFastHTTPServer(env)
	case managed.ServerTypeGin:
		cs.ManagedServer = servers.NewGinServer(env)
	default:
		cs.ManagedServer = servers.NewNetHttpServer(env)
	}
	return cs
}

// Load initial ManagedServer configuration from file
func (cs *ControlServer) LoadServerConfig(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	newConfig := entities.ServerConfig{}
	err = decoder.Decode(&newConfig)
	if err != nil {
		return err
	}
	cs.ManagedServer.SetConfig(newConfig)
	return nil
}
