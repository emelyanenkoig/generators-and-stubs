package server

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

type GinServer struct {
	server *http.Server
}

func (s *GinServer) InitManagedServer(cs *ControlServer) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Создаём маршруты
	cs.createRouter() // Создаём маршрутизатор
	// Переносим маршруты из cs.Config в gin Router

	// Инициализируем сервер gin
	s.server = &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func (s *GinServer) StartManagedServer(cs *ControlServer) {
	log.Println("Managed Server is starting on port 8080 (gin)...")
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on :8080: %v\n", err)
	}
}
