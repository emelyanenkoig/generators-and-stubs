package control

import (
	"encoding/json"
	"gns/stub/env"
	"net/http"
	"os"
	"sync"
	"time"
)

// ControlServer TODO Latency
type ControlServer struct {
	mu            sync.RWMutex
	Config        ServerConfig
	Balancer      Balancer
	ManagedServer ManagedServerInterface
	ControlServer *http.Server
	ReqCount      int
	TpsMu         sync.Mutex
	StartTime     time.Time
	env           env.Environment
}

// ServerConfig defines the structure of the ManagedServer configuration
type ServerConfig struct {
	Paths []ResponsePath `json:"paths"`
}

// ResponsePath defines the structure for each route's response configuration
type ResponsePath struct {
	Path        string      `json:"path"`
	ResponseSet ResponseSet `json:"responseSet"`
}

// ResponseSet defines how to select responses
type ResponseSet struct {
	Choice    string     `json:"choice"` // "round-robin" or "weight"
	Responses []Response `json:"responses"`
}

// Response structure to define each response's properties
type Response struct {
	Weight  int               `json:"weight"`
	Delay   int               `json:"delay"` // in milliseconds
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func NewControlServer(environment env.Environment) *ControlServer {
	cs := &ControlServer{
		Balancer: Balancer{RRobinIndex: map[string]int{}},
	}
	cs.env = environment
	switch environment.ManagedServerType {
	case "fasthttp":
		cs.ManagedServer = &FastHTTPServer{}
	case "gin":
		cs.ManagedServer = &GinServer{}
	default:
		cs.ManagedServer = &NetHttpServer{}
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
	return decoder.Decode(&cs.Config)
}
