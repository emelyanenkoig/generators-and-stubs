package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// GetServerConfig возвращает текущую конфигурацию сервера
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

func (cs *ControlServer) StartControlServer(c *gin.Context) {
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

func (cs *ControlServer) StopControlServer(c *gin.Context) {
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
