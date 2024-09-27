package servers

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"gns/stub/env"
	"gns/stub/log"
	"gns/stub/server/managed/balancing"
	"gns/stub/server/managed/entities"
	"go.uber.org/zap"
	"sync"
	"time"
)

type FastHTTPServer struct {
	Config    entities.ServerConfig
	Balancer  *balancing.Balancer
	server    *fasthttp.Server
	mu        sync.RWMutex
	isRunning bool
	Addr      string
	Port      string
	reqCount  uint
	rpsMu     sync.Mutex
	startTime time.Time
	logger    *zap.Logger
}

func NewFastHTTPServer(env env.Environment) *FastHTTPServer {
	return &FastHTTPServer{
		Addr:     env.ServerAddr,
		Port:     env.ServerPort,
		Balancer: balancing.InitBalancer(),
		logger:   log.InitLogger(env.LogLevel),
	}
}

func (s *FastHTTPServer) InitManagedServer() {
	s.logger.Debug("Initializing managed server (fastHttp)", zap.String("address", s.Addr), zap.String("port", s.Port))
	s.server = &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			s.serverAccessControlMiddlewareFastHTTP(ctx)
		},
		MaxConnsPerIP:      200,
		MaxRequestsPerConn: 200,

		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,
	}
}

func (s *FastHTTPServer) RunManagedServer() {
	s.logger.Info("Running managed server (fastHttp)", zap.String("address", s.Addr), zap.String("port", s.Port))
	s.SetRunning(true)
	if err := s.server.ListenAndServe(fmt.Sprintf(":%s", s.Port)); err != nil {
		s.logger.Fatal("Error starting Gin server", zap.Error(err))
	}
}

func (s *FastHTTPServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

func (s *FastHTTPServer) SetRunning(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v != true {
		s.startTime = time.Time{}
	}
	s.logger.Debug("Managed server running state set to", zap.Bool("running", v))
	s.isRunning = v
}

func (s *FastHTTPServer) GetConfig() entities.ServerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.logger.Debug("Get managed server config", zap.Any("config", s.Config))
	return s.Config
}

func (s *FastHTTPServer) SetConfig(config entities.ServerConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger.Debug("Managed server config is updated", zap.Any("config", config))
	s.Config = config
}

func (s *FastHTTPServer) GetTimeSinceStart() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.startTime
}

func (s *FastHTTPServer) GetReqSinceStart() uint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.reqCount
}

// Middleware для контроля доступа к управляемому серверу
func (s *FastHTTPServer) serverAccessControlMiddlewareFastHTTP(ctx *fasthttp.RequestCtx) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.startTime.IsZero() {
		s.startTime = time.Now()
	}

	if !s.IsRunning() {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		_, _ = ctx.WriteString("Service Unavailable")
		s.logger.Debug("Managed server is not running")
	}

	go func() {
		s.rpsMu.Lock()
		defer s.rpsMu.Unlock()
		s.reqCount++
	}()

	s.RouteHandlerFastHTTP(ctx)
	s.logger.Debug("Managed server is not running",
		zap.ByteString("method", ctx.Request.Header.Method()),
		zap.ByteString("path", ctx.Request.URI().Path()),
		zap.Uint("reqCount", s.reqCount),
	)
}

// Обработчик для управляемого сервера
func (s *FastHTTPServer) RouteHandlerFastHTTP(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, pathConfig := range s.Config.Paths {
		if pathConfig.Path != path {
			continue
		}

		response := s.Balancer.SelectResponse(pathConfig.ResponseSet)
		for key, value := range response.Headers {
			ctx.Response.Header.Set(key, value)
		}
		time.Sleep(time.Duration(response.Delay) * time.Millisecond)
		ctx.SetStatusCode(fasthttp.StatusOK)
		_, err := ctx.WriteString(response.Body)
		if err != nil {
			//log.Printf("Error writing response: %v", err)
		}
		return
	}

	ctx.SetStatusCode(fasthttp.StatusNotFound)
}
