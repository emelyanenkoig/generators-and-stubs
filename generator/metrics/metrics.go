package metrics

import (
	"sync"
	"time"
)

type MetricsCollector struct {
	mu           sync.Mutex
	requests     int
	successes    int
	failures     int
	totalLatency time.Duration
}

func (m *MetricsCollector) GetFailures() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.failures
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

func (mc *MetricsCollector) Record(latency time.Duration, statusCode int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.requests++
	if statusCode >= 200 && statusCode < 300 {
		mc.successes++
	} else {
		mc.failures++
	}
	mc.totalLatency += latency
}

func (mc *MetricsCollector) Report() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	avgLatency := float64(mc.totalLatency.Milliseconds()) / float64(mc.requests)
	successRate := float64(mc.successes) / float64(mc.requests) * 100

	println("Requests:", mc.requests)
	println("Successes:", mc.successes)
	println("Failures:", mc.failures)
	println("Average Latency (ms):", avgLatency)
	println("Success Rate (%):", successRate)
}
