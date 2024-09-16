package main

import (
	"net/http"
	"time"
)

// Middleware to control access to the managed server
func (c *ControlServer) serverAccessControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.mu.RLock()
		defer c.mu.RUnlock()

		if !c.server.running {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		// Перемещаем обработку метрик в асинхронную горутину
		go func() {
			c.tpsMu.Lock()
			defer c.tpsMu.Unlock()

			c.reqCount++
			c.requestTimes = append(c.requestTimes, time.Now())
			c.calculateTPS()
		}()

		next.ServeHTTP(w, r)
	})
}

// Основной метод для расчета TPS и удаления старых метрик
func (c *ControlServer) calculateTPS() {
	now := time.Now()
	if c.lastTpsUpdateTime.IsZero() {
		c.lastTpsUpdateTime = now
		return
	}

	// Удаление старых временных меток
	interval := 10 * time.Second // Рассчитываем TPS за последние 10 секунд
	cutoff := now.Add(-interval)
	recentRequests := c.requestTimes[:0] // Время всех последних запросов за определенный период

	for _, t := range c.requestTimes {
		if t.After(cutoff) {
			recentRequests = append(recentRequests, t)
		}
	}
	c.requestTimes = recentRequests // Обновляем срез времени запросов

	// Рассчитываем TPS
	duration := now.Sub(c.lastTpsUpdateTime).Seconds()
	if duration > 0 {
		c.lastTps = float64(len(recentRequests)) / duration
		c.lastTpsUpdateTime = now
	}
}
