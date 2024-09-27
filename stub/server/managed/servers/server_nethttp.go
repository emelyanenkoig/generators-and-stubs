package servers

import (
	"fmt"
	"github.com/gorilla/mux"
	"gns/stub/env"
	"gns/stub/log"
	"gns/stub/server/managed/balancing"
	"gns/stub/server/managed/entities"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

// реализовать get active conns
type NetHttpServer struct {
	Config    entities.ServerConfig
	Balancer  *balancing.Balancer
	server    *http.Server
	router    *mux.Router
	mu        sync.RWMutex
	isRunning bool
	Addr      string
	Port      string
	reqCount  uint
	rpsMu     sync.Mutex
	startTime time.Time
	logger    *zap.Logger
}

func NewNetHttpServer(env env.Environment) *NetHttpServer {
	return &NetHttpServer{
		Addr:     env.ServerAddr,
		Port:     env.ServerPort,
		Balancer: balancing.InitBalancer(),
		logger:   log.InitLogger(env.LogLevel),
	}
}

func (s *NetHttpServer) InitManagedServer() {
	s.logger.Debug("Initializing managed server (net/http)", zap.String("address", s.Addr), zap.String("port", s.Port))
	s.router = mux.NewRouter()

	for _, pathConfig := range s.Config.Paths {
		s.router.HandleFunc(pathConfig.Path, s.routeHandler).Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete)
	}

	s.router.Use(s.serverAccessControlMiddlewareNetHttp) // middleware

	s.server = &http.Server{
		Addr:           fmt.Sprintf("%s:%s", s.Addr, s.Port),
		Handler:        s.router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func (s *NetHttpServer) RunManagedServer() {
	s.logger.Info("Running managed server (net/http)", zap.String("address", s.Addr), zap.String("port", s.Port))
	s.SetRunning(true)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatal("Error starting net/http server", zap.Error(err))
	}
}

func (s *NetHttpServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

func (s *NetHttpServer) SetRunning(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v == false {
		s.startTime = time.Time{}
	}
	s.logger.Debug("Managed server running state set to", zap.Bool("running", v))
	s.isRunning = v
}

func (s *NetHttpServer) GetConfig() entities.ServerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.logger.Debug("Get managed server config", zap.Any("config", s.Config))
	return s.Config
}

func (s *NetHttpServer) SetConfig(config entities.ServerConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger.Debug("Managed server config is updated", zap.Any("config", config))
	s.Config = config
}

func (s *NetHttpServer) GetTimeSinceStart() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.startTime
}

func (s *NetHttpServer) GetReqSinceStart() uint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.reqCount
}

func (s *NetHttpServer) routeHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	s.mu.RLock() // Lock for reading the configuration
	defer s.mu.RUnlock()

	for _, pathConfig := range s.Config.Paths {
		if pathConfig.Path != path {
			continue
		}
		response := s.Balancer.SelectResponse(pathConfig.ResponseSet) // Select response (round-robin or weighted)
		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}

		time.Sleep(time.Duration(response.Delay) * time.Millisecond)
		_, err := w.Write([]byte(response.Body))
		r.Body.Close()
		if err != nil {
			return
		}
		return
	}

	http.NotFound(w, r)
}

// TODO Сделать middleware тоже структурой с интерфейсом
// Middleware to control access to the managed NetHttpServer
func (s *NetHttpServer) serverAccessControlMiddlewareNetHttp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		if s.startTime.IsZero() {
			s.startTime = time.Now()
		}

		if !s.isRunning {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			s.logger.Debug("Managed server is not running")
			return
		}

		go func() {
			s.rpsMu.Lock()
			defer s.rpsMu.Unlock()
			s.reqCount++
		}()

		next.ServeHTTP(w, r)

		s.logger.Debug("Middleware", zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.Uint("reqCount", s.reqCount))
	})
}
