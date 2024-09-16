package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// ServerConfig defines the structure of the server configuration
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

// TODO либо начинаем отсчет TPS Latency с запуска
type ControlServer struct {
	mu                sync.RWMutex
	config            ServerConfig
	rrIndex           map[string]int
	server            ManagedServer
	controlServer     *http.Server
	reqCount          int
	tpsMu             sync.Mutex
	requestTimes      []time.Time // Добавляем хранение времен запросов
	lastTpsUpdateTime time.Time   // Последнее время обновления TPS
	lastTps           float64     // Последний рассчитанный TPS
}

func NewControlServer() *ControlServer {
	return &ControlServer{
		config: ServerConfig{},
		server: ManagedServer{
			running: true,
		},
		rrIndex: make(map[string]int),
	}
}

// Load initial server configuration from file
func (c *ControlServer) loadServerConfig(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&c.config)
}

type Flags struct {
	responseFilePath  string
	serverPort        string
	controlServerPort string
	addr              string
}

func readFlags() *Flags {
	var f Flags
	f.responseFilePath = os.Getenv("RESPONSE_FILE_PATH")
	f.controlServerPort = os.Getenv("CONTROL_SERVER_PORT")
	f.serverPort = os.Getenv("SERVER_PORT")
	f.addr = os.Getenv("ADDR")

	return &f
}

// TODO добавить логгирование
// + проверка на ошибочный путь
// + чтение из env
// + настройки сервера из env
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	f := readFlags()
	fmt.Println(f)
	cs := NewControlServer()

	// Load the initial server configuration
	err := cs.loadServerConfig(fmt.Sprintf("%s", f.responseFilePath))
	if err != nil {
		log.Fatalf("Error loading server config: %v", err)
	}

	// Set up router for the control server
	controlRouter := mux.NewRouter()

	// Server configuration endpoints
	controlRouter.HandleFunc("/rest/api/v1/server/config", cs.getServerConfigHandler).Methods(http.MethodGet)
	controlRouter.HandleFunc("/rest/api/v1/server/config", cs.updateServerConfigHandler).Methods(http.MethodPost)
	controlRouter.HandleFunc("/rest/api/v1/server/config", cs.deleteServerConfigHandler).Methods(http.MethodDelete)
	controlRouter.HandleFunc("/rest/api/v1/server/start", cs.startServerHandler).Methods(http.MethodPost)
	controlRouter.HandleFunc("/rest/api/v1/server/stop", cs.stopServerHandler).Methods(http.MethodPost)
	controlRouter.HandleFunc("/rest/api/v1/server/status", cs.statusServerHandler).Methods(http.MethodGet)

	// Set up HTTP control server
	cs.controlServer = &http.Server{
		Addr:           ":8062",
		Handler:        controlRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Set up router for the managed server with middleware to control access
	managedRouter := mux.NewRouter()
	managedRouter.Use(cs.serverAccessControlMiddleware) // Middleware to control access based on serverRunning state

	// Register routes dynamically based on the server configuration
	for _, pathConfig := range cs.config.Paths {
		managedRouter.HandleFunc(pathConfig.Path, cs.routeHandler).Methods(http.MethodGet)
	}

	// Set up HTTP managed server
	cs.server.Server = &http.Server{
		Addr:           ":8080",
		Handler:        managedRouter,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start the managed server in a separate goroutine
	go func() {
		log.Println("Managed server is starting on port 8080...")
		if err := cs.server.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on :8081: %v\n", err)
		}
	}()

	// Start the control server
	log.Println("Control server is starting on port 8062...")
	log.Fatal(cs.controlServer.ListenAndServe())
}
