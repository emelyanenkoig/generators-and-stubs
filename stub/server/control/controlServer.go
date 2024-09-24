package control

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

func (cs *ControlServer) InitControlServer() {
	r := gin.Default()

	r.GET("/rest/api/v1/server/config", cs.GetControlServerConfig)
	r.POST("/rest/api/v1/server/config", cs.UpdateControlServerConfig)

	r.DELETE("/rest/api/v1/server/config", cs.DeleteControlServerConfig)
	r.POST("/rest/api/v1/server/start", cs.StartManagedServer)
	r.POST("/rest/api/v1/server/stop", cs.StopManagedServer)
	r.GET("/rest/api/v1/server/status", cs.StatusControlServer)

	cs.ControlServer = &http.Server{
		Addr:           fmt.Sprintf("%s:%s", cs.env.Addr, cs.env.ControlServerPort),
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func (cs *ControlServer) RunControlServer() {
	log.Printf("Starting control server on %s:%s", cs.env.Addr, cs.env.ControlServerPort)
	if err := cs.ControlServer.ListenAndServe(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

}

// GetControlServerConfig return current configuration of ControlServer
func (cs *ControlServer) GetControlServerConfig(c *gin.Context) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	c.JSON(http.StatusOK, cs.Config)
}

func (cs *ControlServer) UpdateControlServerConfig(c *gin.Context) {
	var newConfig ServerConfig
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration"})
		return
	}
	cs.mu.Lock()
	cs.Config = newConfig
	cs.mu.Unlock()

	c.String(http.StatusOK, "The configuration has been successfully applied")
}

func (cs *ControlServer) DeleteControlServerConfig(c *gin.Context) {
	cs.mu.Lock()
	cs.Config = ServerConfig{} // Сброс конфигурации
	cs.mu.Unlock()

	c.String(http.StatusOK, "Configuration deleted successfully")
}

func (cs *ControlServer) StartManagedServer(c *gin.Context) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.ManagedServer.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Server already running"})
		return
	}

	cs.StartTime = time.Now()
	cs.ManagedServer.SetRunning(true)
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

	seconds := time.Since(cs.StartTime).Seconds()

	status := struct {
		Running    bool    `json:"running"`
		TPS        int     `json:"tps"`
		AvgLatency float64 `json:"avg_latency"`
		Duration   float64 `json:"duration"`
	}{
		Running:    cs.ManagedServer.IsRunning(),
		TPS:        cs.ReqCount / int(seconds),
		AvgLatency: 1.0, // Заглушка для средней задержки
		Duration:   seconds,
	}

	c.JSON(http.StatusOK, status)
}

func (cs *ControlServer) InitManagedServer() {
	cs.ManagedServer.InitManagedServer(cs) // Инициализация сервера
}

func (cs *ControlServer) RunManagedServer() {
	cs.ManagedServer.RunManagedServer(cs) // Запуск сервера
}
