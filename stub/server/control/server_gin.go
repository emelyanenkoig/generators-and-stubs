package control

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

type GinServer struct {
	server  *http.Server
	router  *gin.Engine
	running bool
	addr    string
	port    string
}

// Придумать как выпилить ControlServer отсюда
func (s *GinServer) InitManagedServer(cs *ControlServer) {
	gin.SetMode(gin.ReleaseMode)
	s.router = gin.New()

	s.router.Use(ServerAccessControlMiddlewareGin(cs))

	s.router.GET("/", func(c *gin.Context) {
		for _, pathConfig := range cs.Config.Paths {
			if pathConfig.Path != c.Request.URL.Path {
				continue
			}

			response := cs.Balancer.SelectResponse(pathConfig.ResponseSet)
			for key, value := range response.Headers {
				c.Header(key, value)
			}
			time.Sleep(time.Duration(response.Delay) * time.Millisecond)
			c.JSON(http.StatusOK, response.Body)
		}
	})

	s.server = &http.Server{
		Addr:           ":8080",
		Handler:        s.router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func (s *GinServer) RunManagedServer(cs *ControlServer) {
	log.Println("Managed Server is starting on port 8080 (gin)...")
	s.SetRunning(true)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on :8080: %v\n", err)
	}

}

func (s *GinServer) IsRunning() bool {
	return s.running
}

func (s *GinServer) SetRunning(v bool) {
	s.running = v
}

func ServerAccessControlMiddlewareGin(cs *ControlServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		cs.mu.Lock()
		defer cs.mu.Unlock()

		if cs.StartTime.IsZero() {
			cs.StartTime = time.Now()
		}

		if !cs.ManagedServer.IsRunning() {
			c.JSON(http.StatusServiceUnavailable, "Service Unavailable")
			c.Abort()
			return
		}

		go func() {
			cs.TpsMu.Lock()
			defer cs.TpsMu.Unlock()
			cs.ReqCount++
		}()

		c.Next()
	}
}
