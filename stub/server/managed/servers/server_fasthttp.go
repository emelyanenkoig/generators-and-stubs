package servers

import (
	"fmt"
	"github.com/valyala/fasthttp"
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
	proto     string
	certFile  string
	keyFile   string
}

func NewFastHTTPServer(env env.Environment) *FastHTTPServer {
	return &FastHTTPServer{
		Addr:     env.ServerAddr,
		Port:     env.ServerPort,
		Balancer: balancing.InitBalancer(),
		logger:   log.InitLogger(env.LogLevel),
		proto:    env.ProtocolVersion,
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
	s.setProtocolOfServer()
}

func (s *FastHTTPServer) RunManagedServer() {
	s.logger.Info("Running managed server (fastHttp)", zap.String("address", s.Addr), zap.String("port", s.Port))
	s.SetRunning(true)

	switch s.proto {
	case managed.HTTP20:
		// HTTP/2.0 с TLS
		err := fasthttp.ListenAndServeTLS(fmt.Sprintf("%s:%s", s.Addr, s.Port), s.certFile, s.keyFile, s.server.Handler)
		if err != nil {
			s.logger.Fatal("Error starting HTTP/2 server", zap.Error(err))
		}
	case managed.HTTP10:
		// HTTP/1.0 без TLS
		err := fasthttp.ListenAndServe(fmt.Sprintf("%s:%s", s.Addr, s.Port), s.server.Handler)
		if err != nil {
			s.logger.Fatal("Error starting HTTP/1.0 server", zap.Error(err))
		}
	case managed.HTTP11:
		// HTTP/1.1 (обычный режим)
		err := fasthttp.ListenAndServe(fmt.Sprintf("%s:%s", s.Addr, s.Port), s.server.Handler)
		if err != nil {
			s.logger.Fatal("Error starting HTTP/1.1 server", zap.Error(err))
		}
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

func (s *FastHTTPServer) serverAccessControlMiddlewareFastHTTP(ctx *fasthttp.RequestCtx) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.IsRunning() {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		ctx.SetBodyString("Service Unavailable")
		s.logger.Debug("Managed server is not running")
		return
	}

	if err := s.checkRequestProtocolIsValid(ctx); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(err.Error())
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

	s.RouteHandlerFastHTTP(ctx)
}

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
			s.logger.Error("Failed to write response", zap.Error(err))
		}
		return
	}

	ctx.SetStatusCode(fasthttp.StatusNotFound)
	s.logger.Debug("Middleware", zap.ByteString("method", ctx.Method()), zap.ByteString("path", ctx.Path()), zap.Uint("reqCount", s.reqCount))
}

func (s *FastHTTPServer) checkRequestProtocolIsValid(ctx *fasthttp.RequestCtx) error {
	// Проверяем протокол запроса
	proto := string(ctx.Request.Header.Protocol())
	switch s.proto {
	case managed.HTTP10:
		if proto != "HTTP/1.0" {
			return fmt.Errorf("HTTP/1.0 requests only")
		}
		ctx.Response.Header.Set("Connection", "close")
	case managed.HTTP20:
		if proto != "HTTP/2.0" {
			return fmt.Errorf("HTTP/2.0 requests only")
		}
	default:
		if proto != "HTTP/1.1" {
			return fmt.Errorf("HTTP/1.1 requests only")
		}
	}
	return nil
}

func (s *FastHTTPServer) setProtocolOfServer() {
	switch s.proto {
	case managed.HTTP10:
		s.server.DisableKeepalive = true
		s.logger.Info("Using HTTP/1.0 proto")
	case managed.HTTP20:
		s.certFile = "server.crt"
		s.keyFile = "server.key"
		s.logger.Info("Using HTTP/2.0 proto")
	default:
		s.logger.Info("Using default HTTP/1.1 proto")
	}
}

func (s *FastHTTPServer) fastHTTPHandlerWrapper(w http.ResponseWriter, r *http.Request) {
	var ctx fasthttp.RequestCtx
	ctx.Init(&fasthttp.Request{}, nil, nil)
	s.serverAccessControlMiddlewareFastHTTP(&ctx)
	w.Header().Set("Content-Type", "text/plain")
	w.Write(ctx.Response.Body())
}
