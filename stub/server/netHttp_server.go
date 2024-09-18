package server

import (
	"log"
	"net/http"
	"time"
)

type NetHTTPServer struct {
	server    *http.Server
	isRunning bool
}

func (s *NetHTTPServer) InitManagedServer(cs *ControlServer) {
	managedRouter := cs.createRouter()                  // Создаём маршрутизатор
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
