package server

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type GinServer struct {
	cs *ControlServer
}

func NewGinServer(cs *ControlServer) ManagedServerInterface {
	return &GinServer{cs: cs}
}

func (s *GinServer) Init() error {
	router := gin.Default()

	for _, pathConfig := range s.cs.Config.Paths {
		router.Any(pathConfig.Path, func(c *gin.Context) {
			s.cs.RouteHandler(c.Writer, c.Request)
		})
	}

	s.cs.Server.Server = &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	return nil
}

func (s *GinServer) Start() error {
	log.Println("Managed Server is starting on port 8080 (Gin)...")
	if err := s.cs.Server.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
