package control

import (
	"encoding/json"
	"gns/stub/env"
	"net/http"
	"os"
	"sync"
)

// ControlServer TODO Latency
type ControlServer struct {
	ManagedServer ManagedServerInterface
	ControlServer *http.Server
	mu            sync.RWMutex
	addr          string
	port          string
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

func NewControlServer(env env.Environment) *ControlServer {
	cs := &ControlServer{
		addr: env.ControlServerAddr,
		port: env.ControlServerPort,
	}
	switch env.ManagedServerType {
	//case "fasthttp":
	//	cs.ManagedServer = &FastHTTPServer{
	//		addr: env.ServerAddr,
	//		port: env.ServerPort,
	//	}
	case "gin":
		cs.ManagedServer = &GinServer{
			addr:     env.ServerAddr,
			port:     env.ServerPort,
			Balancer: Balancer{make(map[string]int)},
		}
	default:
		cs.ManagedServer = &NetHttpServer{
			addr:     env.ServerAddr,
			port:     env.ServerPort,
			Balancer: Balancer{RRobinIndex: make(map[string]int)},
		}
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
	newConfig := ServerConfig{}
	err = decoder.Decode(&newConfig)
	if err != nil {
		return err
	}
	cs.ManagedServer.SetConfig(newConfig)
	return nil
}
