package server

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"
)

// ControlServer TODO либо начинаем отсчет TPS Latency с запуска
type ControlServer struct {
	mu            sync.RWMutex
	Config        ServerConfig
	RRobinIndex   map[string]int
	Server        ManagedServer
	ControlServer *http.Server
	ReqCount      int
	TpsMu         sync.Mutex
	StartTime     time.Time
}

func NewControlServer() *ControlServer {
	return &ControlServer{
		Config:      ServerConfig{},
		RRobinIndex: make(map[string]int),
	}
}

// Load initial Server configuration from file
func (c *ControlServer) LoadServerConfig(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&c.Config)
}
