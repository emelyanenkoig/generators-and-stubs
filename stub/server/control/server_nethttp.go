package control

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type NetHttpServer struct {
	server    *http.Server
	router    *mux.Router
	isRunning bool
	addr      string
	port      string
}

func (s *NetHttpServer) InitManagedServer(cs *ControlServer) {
	s.router = mux.NewRouter()

	for _, pathConfig := range cs.Config.Paths {
		s.router.HandleFunc(pathConfig.Path, cs.RouteHandler).Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete)
	}

	s.router.Use(cs.ServerAccessControlMiddlewareNetHttp) // middleware

	s.server = &http.Server{
		Addr:           ":8080",
		Handler:        s.router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func (s *NetHttpServer) RunManagedServer(cs *ControlServer) {
	log.Println("Managed Server is starting on port 8080 (net/http)...")
	s.SetRunning(true)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on :8080: %v\n", err)
	}
}

func (s *NetHttpServer) IsRunning() bool {
	return s.isRunning
}

func (s *NetHttpServer) SetRunning(v bool) {
	s.isRunning = v
}

// Middleware to control access to the managed NetHttpServer
func (cs *ControlServer) ServerAccessControlMiddlewareNetHttp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cs.mu.RLock()
		defer cs.mu.RUnlock()

		if cs.StartTime.IsZero() {
			cs.StartTime = time.Now()
		}

		if !cs.ManagedServer.IsRunning() {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		// Перемещаем обработку метрик в асинхронную горутину
		go func() {
			cs.TpsMu.Lock()
			defer cs.TpsMu.Unlock()
			cs.ReqCount++
		}()

		next.ServeHTTP(w, r)
	})
}

// Handler for processing requests on configured routes
func (cs *ControlServer) RouteHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	cs.mu.RLock() // Lock for reading the configuration
	defer cs.mu.RUnlock()

	for _, pathConfig := range cs.Config.Paths {
		if pathConfig.Path != path {
			continue
		}
		response := cs.Balancer.SelectResponse(pathConfig.ResponseSet) // Select response (round-robin or weighted)
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
