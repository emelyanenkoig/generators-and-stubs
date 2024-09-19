package server

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type NetHTTPServer struct {
	server    *http.Server
	isRunning bool
}

func (s *NetHTTPServer) InitManagedServer(cs *ControlServer) {
	managedRouter := mux.NewRouter()

	// Добавляем маршруты на основе конфигурации сервера
	for _, pathConfig := range cs.Config.Paths {
		managedRouter.HandleFunc(pathConfig.Path, cs.RouteHandler).Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete)
	}

	managedRouter.Use(cs.ServerAccessControlMiddleware) // Добавляем middleware

	// Инициализируем сервер net/http
	s.server = &http.Server{
		Addr:           ":8080",
		Handler:        managedRouter,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func (s *NetHTTPServer) StartManagedServer(cs *ControlServer) {
	log.Println("Managed Server is starting on port 8080 (net/http)...")
	s.isRunning = true
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on :8080: %v\n", err)
	}
}

func (s *NetHTTPServer) IsRunning() bool {
	return s.isRunning
}

func (s *NetHTTPServer) SetRunning(v bool) {
	s.isRunning = v
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
