package control

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sync"
	"time"
)

type NetHttpServer struct {
	Config    ServerConfig
	Balancer  Balancer
	server    *http.Server
	router    *mux.Router
	mu        sync.RWMutex
	isRunning bool
	addr      string
	port      string
	reqCount  uint
	rpsMu     sync.Mutex
	startTime time.Time
}

func (s *NetHttpServer) InitManagedServer() {
	s.router = mux.NewRouter()

	for _, pathConfig := range s.Config.Paths {
		s.router.HandleFunc(pathConfig.Path, s.routeHandler).Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete)
	}

	s.router.Use(s.serverAccessControlMiddlewareNetHttp) // middleware

	s.server = &http.Server{
		Addr:           fmt.Sprintf("%s:%s", s.addr, s.port),
		Handler:        s.router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func (s *NetHttpServer) RunManagedServer() {
	log.Printf("Managed Server is starting on port %s (net/http)...", s.port)
	s.SetRunning(true)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on :%s: %v\n", s.port, err)
	}
}

func (s *NetHttpServer) IsRunning() bool {
	return s.isRunning
}

func (s *NetHttpServer) SetRunning(v bool) {
	if v == false {
		s.startTime = time.Time{}
	}
	s.isRunning = v
}

func (s *NetHttpServer) GetConfig() ServerConfig {
	return s.Config
}

func (s *NetHttpServer) SetConfig(config ServerConfig) {
	s.Config = config
}

func (s *NetHttpServer) GetTimeSinceStart() time.Time {
	return s.startTime
}

func (s *NetHttpServer) GetReqSinceStart() uint {
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
			return
		}

		go func() {
			s.rpsMu.Lock()
			defer s.rpsMu.Unlock()
			s.reqCount++
		}()
		next.ServeHTTP(w, r)
	})
}
