package server

import (
	"encoding/json"
	"net/http"
	"time"
)

// HTTP handler to get the current ManagedServer configuration
func (cs *ControlServer) GetServerConfigHandler(w http.ResponseWriter, r *http.Request) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(cs.Config)
	if err != nil {
		return
	}
}

// HTTP handler to update the ManagedServer configuration
func (cs *ControlServer) UpdateServerConfigHandler(w http.ResponseWriter, r *http.Request) {
	var newConfig ServerConfig
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newConfig)
	if err != nil {
		http.Error(w, "Invalid configuration", http.StatusBadRequest)
		return
	}

	cs.mu.Lock()
	cs.Config = newConfig
	cs.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("The configuration has been successfully applied"))
}

// HTTP handler to delete the ManagedServer configuration
func (cs *ControlServer) DeleteServerConfigHandler(w http.ResponseWriter, r *http.Request) {
	cs.mu.Lock()
	cs.Config = ServerConfig{} // Reset to default empty Config
	cs.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Configuration deleted successfully"))
}

// HTTP handler to start the managed ManagedServer (enabling request handling)
func (cs *ControlServer) StartServerHandler(w http.ResponseWriter, r *http.Request) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.ManagedServer.IsRunning() {
		http.Error(w, "ManagedServer already running", http.StatusBadRequest)
		return
	}

	cs.StartTime = time.Now()
	cs.ManagedServer.SetRunning(true)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ManagedServer started successfully"))
}

// HTTP handler to stop the managed ManagedServer (disabling request handling)
func (cs *ControlServer) StopServerHandler(w http.ResponseWriter, r *http.Request) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if !cs.ManagedServer.IsRunning() {
		http.Error(w, "ManagedServer is not running", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	cs.ManagedServer.SetRunning(false)
	cs.ReqCount = 0
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ManagedServer stopped successfully"))
}

func (cs *ControlServer) StatusServerHandler(w http.ResponseWriter, r *http.Request) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	seconds := time.Since(cs.StartTime).Seconds()

	cs.TpsMu.Lock()
	defer cs.TpsMu.Unlock()
	tps := cs.ReqCount / int(seconds)

	// Формируем ответ в JSON формате
	status := struct {
		Running    bool    `json:"running"`
		TPS        int     `json:"tps"`
		AvgLatency float64 `json:"avg_latency"` // В миллисекундах
		Duruation  float64 `json:"druation"`
	}{
		Running:    cs.ManagedServer.IsRunning(),
		TPS:        tps,
		AvgLatency: 1,
		Duruation:  float64(seconds),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Handler for processing requests on configured routes
func (cs *ControlServer) RouteHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	cs.mu.RLock() // Lock for reading the configuration
	defer cs.mu.RUnlock()

	for _, pathConfig := range cs.Config.Paths {
		if pathConfig.Path == path {
			response := cs.SelectResponse(pathConfig.ResponseSet) // Select response (round-robin or weighted)
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
