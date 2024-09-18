package server

import (
	"net/http"
	"time"
)

// Middleware to control access to the managed ManagedServer
func (cs *ControlServer) ServerAccessControlMiddleware(next http.Handler) http.Handler {
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
