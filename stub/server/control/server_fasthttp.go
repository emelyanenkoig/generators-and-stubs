package control

import (
	"github.com/valyala/fasthttp"
	"log"
	"time"
)

type FastHTTPServer struct {
	server  *fasthttp.Server
	running bool
	addr    string
	port    string
}

func (s *FastHTTPServer) InitManagedServer(cs *ControlServer) {
	s.server = &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			cs.ServerAccessControlMiddlewareFastHTTP(ctx)
		},
		MaxConnsPerIP:      200,
		MaxRequestsPerConn: 200,

		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,
	}
}

func (s *FastHTTPServer) RunManagedServer(cs *ControlServer) {
	log.Println("Managed Server is starting on port 8080 (fasthttp)...")
	s.SetRunning(true)

	if err := s.server.ListenAndServe(":8080"); err != nil {
		log.Fatalf("Could not listen on :8080: %v\n", err)
	}
}

func (s *FastHTTPServer) IsRunning() bool {
	return s.running
}

func (s *FastHTTPServer) SetRunning(v bool) {
	s.running = v
}

// Middleware для контроля доступа к управляемому серверу
func (cs *ControlServer) ServerAccessControlMiddlewareFastHTTP(ctx *fasthttp.RequestCtx) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if cs.StartTime.IsZero() {
		cs.StartTime = time.Now()
	}

	if !cs.ManagedServer.IsRunning() {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		_, _ = ctx.WriteString("Service Unavailable")
		return
	}

	// Обновляем счетчик TPS
	cs.TpsMu.Lock()
	cs.ReqCount++
	cs.TpsMu.Unlock()

	// Здесь вызываем обработчик
	cs.RouteHandlerFastHTTP(ctx)
}

// Обработчик для управляемого сервера
func (cs *ControlServer) RouteHandlerFastHTTP(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	for _, pathConfig := range cs.Config.Paths {
		if pathConfig.Path != path {
			continue
		}

		response := cs.Balancer.SelectResponse(pathConfig.ResponseSet)
		for key, value := range response.Headers {
			ctx.Response.Header.Set(key, value)
		}
		time.Sleep(time.Duration(response.Delay) * time.Millisecond)
		ctx.SetStatusCode(fasthttp.StatusOK)
		_, err := ctx.WriteString(response.Body)
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}
		return
	}

	ctx.SetStatusCode(fasthttp.StatusNotFound)
}
