package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// HTTP handler to get the current server configuration
func (c *ControlServer) getServerConfigHandler(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(c.config)
	if err != nil {
		return
	}
}

// HTTP handler to update the server configuration
func (c *ControlServer) updateServerConfigHandler(w http.ResponseWriter, r *http.Request) {
	var newConfig ServerConfig
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newConfig)
	if err != nil {
		http.Error(w, "Invalid configuration", http.StatusBadRequest)
		return
	}

	c.mu.Lock()
	c.config = newConfig
	c.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("The configuration has been successfully applied"))
}

// HTTP handler to delete the server configuration
func (c *ControlServer) deleteServerConfigHandler(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	c.config = ServerConfig{} // Reset to default empty config
	c.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Configuration deleted successfully"))
}

// HTTP handler to start the managed server (enabling request handling)
func (c *ControlServer) startServerHandler(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.server.running {
		http.Error(w, "Server already running", http.StatusBadRequest)
		return
	}

	c.server.running = true
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server started successfully"))
}

// HTTP handler to stop the managed server (disabling request handling)
func (c *ControlServer) stopServerHandler(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.server.running {
		http.Error(w, "Server is not running", http.StatusBadRequest)
		return
	}

	c.server.running = false
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server stopped successfully"))
}

func (c *ControlServer) statusServerHandler(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Рассчитываем среднюю задержку (latency)
	var totalLatency time.Duration
	for i := 1; i < len(c.requestTimes); i++ {
		totalLatency += c.requestTimes[i].Sub(c.requestTimes[i-1])
	}
	var avgLatency float64
	if len(c.requestTimes) > 1 {
		avgLatency = float64(totalLatency.Milliseconds()) / float64(len(c.requestTimes)-1)
	}

	// Формируем ответ в JSON формате
	status := struct {
		Running    bool    `json:"running"`
		TPS        float64 `json:"tps"`
		AvgLatency float64 `json:"avg_latency"` // В миллисекундах
	}{
		Running:    c.server.running,
		TPS:        c.lastTps,
		AvgLatency: avgLatency,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Handler for processing requests on configured routes
func (c *ControlServer) routeHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	c.mu.RLock() // Lock for reading the configuration
	defer c.mu.RUnlock()

	for _, pathConfig := range c.config.Paths {
		if pathConfig.Path == path {
			response := c.selectResponse(pathConfig.ResponseSet) // Select response (round-robin or weighted)
			for key, value := range response.Headers {
				w.Header().Set(key, value)
			}
			time.Sleep(time.Duration(response.Delay) * time.Millisecond) // Simulate response delay
			w.Write([]byte(response.Body))                               // Write response body
			return
		}
	}

	http.NotFound(w, r)
}
