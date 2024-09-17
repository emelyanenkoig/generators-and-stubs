package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gns/stub/env"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// ServerConfig defines the structure of the Server configuration
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

type ManagedServer struct {
	*http.Server
	running bool
}

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
		Config: ServerConfig{},
		Server: ManagedServer{
			running: true,
		},
		RRobinIndex: make(map[string]int),
	}
}

func (cs *ControlServer) InitControlServer(f env.Environment) {

	// Set up router for the control Server
	controlRouter := mux.NewRouter()

	// Server configuration endpoints
	controlRouter.HandleFunc("/rest/api/v1/server/config", cs.UpdateServerConfigHandler).Methods(http.MethodPost)
	controlRouter.HandleFunc("/rest/api/v1/server/config", cs.DeleteServerConfigHandler).Methods(http.MethodDelete)
	controlRouter.HandleFunc("/rest/api/v1/server/start", cs.StartServerHandler).Methods(http.MethodPost)
	controlRouter.HandleFunc("/rest/api/v1/server/stop", cs.StopServerHandler).Methods(http.MethodPost)
	controlRouter.HandleFunc("/rest/api/v1/server/status", cs.StatusServerHandler).Methods(http.MethodGet)

	// Set up HTTP control Server
	cs.ControlServer = &http.Server{
		Addr:           fmt.Sprintf(":%s", f.ControlServerPort),
		Handler:        controlRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start the control Server
	log.Println("Control Server is starting on port 8062...")
	log.Fatal(cs.ControlServer.ListenAndServe())

}

func (cs *ControlServer) InitManagedServer() {

	// Set up router for the managed Server with middleware to control access
	managedRouter := mux.NewRouter()
	managedRouter.Use(cs.ServerAccessControlMiddleware) // Middleware to control access based on serverRunning state

	// Register routes dynamically based on the Server configuration
	// TODO потестить что хендлеры читает метод
	for _, pathConfig := range cs.Config.Paths {
		managedRouter.HandleFunc(pathConfig.Path, cs.RouteHandler).Methods(http.MethodGet)
		managedRouter.HandleFunc(pathConfig.Path, cs.RouteHandler).Methods(http.MethodPost)
		managedRouter.HandleFunc(pathConfig.Path, cs.RouteHandler).Methods(http.MethodDelete)
		managedRouter.HandleFunc(pathConfig.Path, cs.RouteHandler).Methods(http.MethodPut)
	}

	// Set up HTTP managed Server
	cs.Server.Server = &http.Server{
		Addr:           ":8080",
		Handler:        managedRouter,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("Managed Server is starting on port 8080...")
	if err := cs.Server.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on :8081: %v\n", err)
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
