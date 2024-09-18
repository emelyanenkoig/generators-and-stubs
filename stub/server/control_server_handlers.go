package server

import (
	"encoding/json"
	"net/http"
	"time"
)

// GetServerConfigHandler - получить текущую конфигурацию сервера
func (c *ControlServer) GetServerConfigHandler(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(c.Config)
	if err != nil {
		return
	}
}

// StartServerHandler - запуск управляемого сервера
func (c *ControlServer) StartServerHandler(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Server.running {
		http.Error(w, "Server already running", http.StatusBadRequest)
		return
	}

	c.StartTime = time.Now()
	c.Server.running = true
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server started successfully"))
}
