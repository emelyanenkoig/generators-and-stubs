package server

import (
	"net/http"
	"time"
)

// Middleware to control access to the managed Server
func (c *ControlServer) ServerAccessControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.mu.RLock()
		defer c.mu.RUnlock()

		if c.StartTime.IsZero() {
			c.StartTime = time.Now()
		}

		if !c.Server.running {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		// Перемещаем обработку метрик в асинхронную горутину
		go func() {
			c.TpsMu.Lock()
			defer c.TpsMu.Unlock()
			c.ReqCount++
		}()

		next.ServeHTTP(w, r)
	})
}
