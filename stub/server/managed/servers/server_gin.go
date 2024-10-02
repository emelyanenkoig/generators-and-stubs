package servers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gns/stub/env"
	"gns/stub/log"
	"gns/stub/server/managed"
	"gns/stub/server/managed/balancing"
	"gns/stub/server/managed/entities"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

type GinServer struct {
	Config    entities.ServerConfig
	Balancer  *balancing.Balancer
	server    *http.Server
	router    *gin.Engine
	mu        sync.RWMutex
	isRunning bool
	Addr      string
	Port      string
	reqCount  uint
	rpsMu     sync.Mutex
	startTime time.Time
	logger    *zap.Logger
	proto     string
	certFile  string
	keyFile   string
}

func NewGinServer(env env.Environment) *GinServer {
	return &GinServer{
		Addr:     env.ServerAddr,
		Port:     env.ServerPort,
		Balancer: balancing.InitBalancer(),
		logger:   log.InitLogger(env.LogLevel),
		proto:    env.ProtocolVersion,
	}
}

func (s *GinServer) InitManagedServer() {
	s.logger.Debug("Initializing managed server (Gin)", zap.String("address", s.Addr), zap.String("port", s.Port))
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
			response := s.Balancer.SelectResponse(pathConfig.ResponseSet)
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
		Addr:           fmt.Sprintf("%s:%s", s.Addr, s.Port),
		Handler:        s.router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.setProtocolOfServer()
}

func (s *GinServer) RunManagedServer() {
	s.logger.Info("Running managed server (Gin)", zap.String("address", s.Addr), zap.String("port", s.Port))
	s.SetRunning(true)

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
	s.logger.Debug("Managed server running state set to", zap.Bool("running", v))
	s.isRunning = v
}

func (s *GinServer) GetConfig() entities.ServerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.logger.Debug("Get managed server config", zap.Any("config", s.Config))
	return s.Config
}

func (s *GinServer) SetConfig(config entities.ServerConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger.Debug("Managed server config is updated", zap.Any("config", config))
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
		defer s.mu.RUnlock()

		if !s.isRunning {
			c.JSON(http.StatusServiceUnavailable, "Service Unavailable")
			c.Abort()
			s.logger.Debug("Managed server is not running")
			return
		}

		if err := s.checkRequestProtocolIsValid(c); err != nil {
			c.JSON(http.StatusBadRequest, "Invalid proto of request")
			s.logger.Error("Invalid proto of request", zap.Error(err))
			return
		}

		if s.startTime.IsZero() {
			s.startTime = time.Now()
		}

		go func() {
			s.rpsMu.Lock()
			defer s.rpsMu.Unlock()
			s.reqCount++
		}()
		c.Next()

		s.logger.Debug("Middleware", zap.String("method", c.Request.Method), zap.String("path", c.Request.URL.Path), zap.Uint("reqCount", s.reqCount))
	}
}

func (s *GinServer) checkRequestProtocolIsValid(c *gin.Context) error {
	switch s.proto {
	case managed.HTTP10:
		if c.Request.ProtoMajor != 1 || c.Request.ProtoMinor != 0 {
			return fmt.Errorf("HTTP/1.0 requests only")
		}
	case managed.HTTP20:
		if c.Request.ProtoMajor != 2 || c.Request.ProtoMinor != 0 {
			return fmt.Errorf("HTTP/2.0 requests only")
		}
	default:
		if c.Request.ProtoMajor != 1 || c.Request.ProtoMinor != 1 {
			return fmt.Errorf("HTTP/1.1 requests only")
		}
	}
	return nil

}

func (s *GinServer) setProtocolOfServer() {
	switch s.proto {
	case managed.HTTP10:
		s.server.SetKeepAlivesEnabled(false)
		s.logger.Info("Using HTTP/1.0 proto")

	case managed.HTTP20:
		s.certFile = "server.crt"
		s.keyFile = "server.key"
		s.logger.Info("Using HTTP/2.0 proto")
	default:
		s.logger.Info("Using HTTP/1.1 proto")
	}
}
