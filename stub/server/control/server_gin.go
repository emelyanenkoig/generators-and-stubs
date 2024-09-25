package control

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync"
	"time"
)

type GinServer struct {
	Config    ServerConfig
	Balancer  Balancer
	server    *http.Server
	router    *gin.Engine
	mu        sync.RWMutex
	isRunning bool
	addr      string
	port      string
	reqCount  uint
	rpsMu     sync.Mutex
	startTime time.Time
}

func (s *GinServer) InitManagedServer() {
	gin.SetMode(gin.ReleaseMode)
	s.router = gin.New()

	s.router.Use(s.serverAccessControlMiddlewareGin())

	s.router.GET("/", func(c *gin.Context) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		for _, pathConfig := range s.Config.Paths {
			if pathConfig.Path != c.Request.URL.Path {
				continue
			}
			// TODO разобраться с race condition
			s.mu.Lock()
			response := s.Balancer.SelectResponse(pathConfig.ResponseSet)
			s.mu.Unlock()
			for key, value := range response.Headers {
				c.Header(key, value)
			}
			time.Sleep(time.Duration(response.Delay) * time.Millisecond)
			c.JSON(http.StatusOK, response.Body)
			return
		}

		c.JSON(http.StatusNotFound, "Path not found")
	})

	s.server = &http.Server{
		Addr:           fmt.Sprintf("%s:%s", s.addr, s.port),
		Handler:        s.router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func (s *GinServer) RunManagedServer() {
	log.Printf("Managed Server is starting on port %s (gin)...", s.port)
	s.SetRunning(true)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on :%s: %v\n", s.port, err)
	}

}

func (s *GinServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

func (s *GinServer) SetRunning(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v == false {
		s.startTime = time.Time{}
	}
	s.isRunning = v
}

func (s *GinServer) GetConfig() ServerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Config
}

func (s *GinServer) SetConfig(config ServerConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Config = config
}

func (s *GinServer) GetTimeSinceStart() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.startTime
}

func (s *GinServer) GetReqSinceStart() uint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.reqCount
}

func (s *GinServer) serverAccessControlMiddlewareGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		s.mu.RLock()
		if !s.isRunning {
			s.mu.RUnlock()
			c.JSON(http.StatusServiceUnavailable, "Service Unavailable")
			c.Abort()
			return
		}
		s.mu.RUnlock()

		s.rpsMu.Lock()
		s.reqCount++
		s.rpsMu.Unlock()

		c.Next()
	}
}
