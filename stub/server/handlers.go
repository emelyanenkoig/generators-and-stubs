package server

//
//import (
//	"encoding/json"
//	"net/http"
//	"time"
//)
//
//// HTTP handler to get the current Server configuration
//func (c *ControlServer) GetServerConfigHandler(w http.ResponseWriter, r *http.Request) {
//	c.mu.RLock()
//	defer c.mu.RUnlock()
//
//	w.Header().Set("Content-Type", "application/json")
//	err := json.NewEncoder(w).Encode(c.Config)
//	if err != nil {
//		return
//	}
//}
//
//// HTTP handler to update the Server configuration
//func (c *ControlServer) UpdateServerConfigHandler(w http.ResponseWriter, r *http.Request) {
//	var newConfig ServerConfig
//	decoder := json.NewDecoder(r.Body)
//	err := decoder.Decode(&newConfig)
//	if err != nil {
//		http.Error(w, "Invalid configuration", http.StatusBadRequest)
//		return
//	}
//
//	c.mu.Lock()
//	c.Config = newConfig
//	c.mu.Unlock()
//
//	w.WriteHeader(http.StatusOK)
//	w.Write([]byte("The configuration has been successfully applied"))
//}
//
//// HTTP handler to delete the Server configuration
//func (c *ControlServer) DeleteServerConfigHandler(w http.ResponseWriter, r *http.Request) {
//	c.mu.Lock()
//	c.Config = ServerConfig{} // Reset to default empty Config
//	c.mu.Unlock()
//
//	w.WriteHeader(http.StatusOK)
//	w.Write([]byte("Configuration deleted successfully"))
//}
//
//// HTTP handler to start the managed Server (enabling request handling)
//func (c *ControlServer) StartServerHandler(w http.ResponseWriter, r *http.Request) {
//	c.mu.Lock()
//	defer c.mu.Unlock()
//
//	if c.Server.running {
//		http.Error(w, "Server already running", http.StatusBadRequest)
//		return
//	}
//
//	c.StartTime = time.Now()
//	c.Server.running = true
//	w.WriteHeader(http.StatusOK)
//	w.Write([]byte("Server started successfully"))
//}
//
//// HTTP handler to stop the managed Server (disabling request handling)
//func (c *ControlServer) StopServerHandler(w http.ResponseWriter, r *http.Request) {
//	c.mu.Lock()
//	defer c.mu.Unlock()
//
//	if !c.Server.running {
//		http.Error(w, "Server is not running", http.StatusBadRequest)
//		return
//	}
//	defer r.Body.Close()
//	c.Server.running = false
//	c.ReqCount = 0
//	w.WriteHeader(http.StatusOK)
//	w.Write([]byte("Server stopped successfully"))
//}
//
//func (c *ControlServer) StatusServerHandler(w http.ResponseWriter, r *http.Request) {
//	c.mu.RLock()
//	defer c.mu.RUnlock()
//
//	seconds := time.Since(c.StartTime).Seconds()
//
//	c.TpsMu.Lock()
//	defer c.TpsMu.Unlock()
//	tps := c.ReqCount / int(seconds)
//
//	// Формируем ответ в JSON формате
//	status := struct {
//		Running    bool    `json:"running"`
//		TPS        int     `json:"tps"`
//		AvgLatency float64 `json:"avg_latency"` // В миллисекундах
//		Duruation  float64 `json:"druation"`
//	}{
//		Running:    c.Server.running,
//		TPS:        tps,
//		AvgLatency: 1,
//		Duruation:  float64(seconds),
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	json.NewEncoder(w).Encode(status)
//}
//
//// Handler for processing requests on configured routes
//func (c *ControlServer) RouteHandler(w http.ResponseWriter, r *http.Request) {
//	path := r.URL.Path
//	c.mu.RLock() // Lock for reading the configuration
//	defer c.mu.RUnlock()
//
//	for _, pathConfig := range c.Config.Paths {
//		if pathConfig.Path == path {
//			response := c.SelectResponse(pathConfig.ResponseSet) // Select response (round-robin or weighted)
//			for key, value := range response.Headers {
//				w.Header().Set(key, value)
//			}
//			time.Sleep(time.Duration(response.Delay) * time.Millisecond) // Simulate response delay
//			w.Write([]byte(response.Body))                               // Write response body
//			return
//		}
//	}
//
//	http.NotFound(w, r)
//}
