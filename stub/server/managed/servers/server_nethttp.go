package servers

import (
	"fmt"
	"github.com/gorilla/mux"
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
	proto     string
	certFile  string
	keyFile   string
}

func NewNetHttpServer(env env.Environment) *NetHttpServer {
	return &NetHttpServer{
		Addr:     env.ServerAddr,
		Port:     env.ServerPort,
		Balancer: balancing.InitBalancer(),
		logger:   log.InitLogger(env.LogLevel),
		proto:    env.ProtocolVersion,
	}
}

func (s *NetHttpServer) InitManagedServer() {
	s.logger.Debug("Initializing managed server (net/http)", zap.String("address", s.Addr), zap.String("port", s.Port))
	s.router = mux.NewRouter()

	for _, pathConfig := range s.Config.Paths {
		s.logger.Debug("Setting route", zap.String("path", pathConfig.Path))
		s.router.HandleFunc(pathConfig.Path, s.routeHandler).
			Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete)
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
	s.setProtocolOfServer()

}

func (s *NetHttpServer) RunManagedServer() {
	s.logger.Info("Running managed server (net/http)", zap.String("address", s.Addr), zap.String("port", s.Port), zap.String("Server protocol", s.proto))
	s.SetRunning(true)

	switch s.proto {
	case managed.HTTP20:
		err := s.server.ListenAndServeTLS(s.certFile, s.keyFile)
		if err != nil {
			s.logger.Fatal("Error starting net/http server", zap.Error(err))
		}
	case managed.HTTP10:
		err := s.server.ListenAndServe()
		if err != nil {
			s.logger.Fatal("Error starting net/http server", zap.Error(err))
		}
	default:
		err := s.server.ListenAndServe()
		if err != nil {
			s.logger.Fatal("Error starting net/http server", zap.Error(err))
		}
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

func (s *NetHttpServer) SetConfig(config entities.ServerConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isValidConfigChoice(&config) {
		s.Config = config
		s.UpdateRoutes()
		s.logger.Debug("Managed server config is updated", zap.Any("config", config))
		return nil
	}
	s.logger.Debug("Failed to update config, invalid config", zap.Any("config", config))
	return fmt.Errorf("invalid config")

}

func (s *NetHttpServer) isValidConfigChoice(c *entities.ServerConfig) bool {

	for _, path := range c.Paths {
		switch path.ResponseSet.Choice {
		case balancing.Weighted:
			continue
		case balancing.Random:
			continue
		case balancing.RoundRobin:
			continue
		case balancing.WeightedRandomWithBinarySearch:
			continue
		default:
			return false
		}
	}
	return true

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
		err, response := s.Balancer.SelectResponse(pathConfig.ResponseSet) // Select response (round-robin or weighted)
		if err != nil {
			http.Error(w, "invalid choice", http.StatusBadRequest)
			s.logger.Error("invalid choice", zap.Error(err))
			return
		}
		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}

		time.Sleep(time.Duration(response.Delay) * time.Millisecond)
		_, err = w.Write([]byte(response.Body))
		r.Body.Close()
		if err != nil {
			return
		}
		return
	}

	http.NotFound(w, r)
}

// Middleware to control access to the managed NetHttpServer
func (s *NetHttpServer) serverAccessControlMiddlewareNetHttp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		if !s.isRunning {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			s.logger.Debug("Managed server is not running")
			return
		}

		if err := s.checkRequestProtocolIsValid(r); err != nil {
			http.Error(w, "Invalid proto of request", http.StatusBadRequest)
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

		next.ServeHTTP(w, r)

		s.logger.Debug("Middleware", zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.Uint("reqCount", s.reqCount))
	})
}

func (s *NetHttpServer) checkRequestProtocolIsValid(r *http.Request) error {
	switch s.proto {
	case managed.HTTP10:
		if r.ProtoMajor != 1 || r.ProtoMinor != 0 {
			return fmt.Errorf("HTTP/1.0 requests only")
		}
	case managed.HTTP20:
		if r.ProtoMajor != 2 || r.ProtoMinor != 0 {
			return fmt.Errorf("HTTP/2.0 requests only")
		}
	default:
		if r.ProtoMajor != 1 || r.ProtoMinor != 1 {
			return fmt.Errorf("HTTP/1.1 requests only")
		}
	}
	return nil

}

func (s *NetHttpServer) setProtocolOfServer() {
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

func (s *NetHttpServer) UpdateRoutes() {
	// Очистка текущих маршрутов
	s.router = mux.NewRouter()

	// Добавление новых маршрутов из обновленной конфигурации
	for _, pathConfig := range s.Config.Paths {
		s.logger.Debug("Setting updated route", zap.String("path", pathConfig.Path))
		s.router.HandleFunc(pathConfig.Path, s.routeHandler).
			Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete)
	}

	// Повторное добавление middleware
	s.router.Use(s.serverAccessControlMiddlewareNetHttp)

	// Применение обновленного роутера
	s.server.Handler = s.router
	s.logger.Info("Routes updated successfully")
}
