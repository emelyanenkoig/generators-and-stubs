package loadgen

import (
	"context"
	"fmt"
	c "gns/generator/config"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type LoadGenerator struct {
	config       *c.Config
	mu           sync.Mutex
	countFailure int
	countSuccess int
}

func NewLoadGenerator(config *c.Config) *LoadGenerator {
	return &LoadGenerator{config: config}
}

func (lg *LoadGenerator) Start() {
	for _, reqConfig := range lg.config.Requests {
		lg.runLoadTest(reqConfig)
	}
}

func (lg *LoadGenerator) runLoadTest(reqConfig c.RequestConfig) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(reqConfig.Duration)*time.Second)
	defer cancel()

	for i := 0; i < reqConfig.Threads; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			lg.sendRequests(ctx, reqConfig)
		}(i)
	}

	wg.Wait()
	fmt.Printf("\nCount of Failed: %v\n", lg.countFailure)
	fmt.Printf("\nCount of Success: %v\n", lg.countSuccess)
}

func (lg *LoadGenerator) sendRequests(ctx context.Context, reqConfig c.RequestConfig) {
	transport := &http.Transport{
		MaxIdleConnsPerHost: 50000,
		MaxIdleConns:        50000,
		IdleConnTimeout:     100 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
	req, err := http.NewRequest(reqConfig.Method, reqConfig.URL, nil) // Add body if needed
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return
	}

	for k, v := range reqConfig.Headers {
		req.Header.Set(k, v)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			start := time.Now()
			resp, err := client.Do(req)
			defer resp.Body.Close()
			_, _ = io.ReadAll(resp.Body)
			duration := time.Since(start)

			if err != nil {
				log.Printf("Request failed: %v\n", err)
				lg.mu.Lock()
				lg.countFailure++
				lg.mu.Unlock()
			} else {
				log.Printf("Response Status: %s, Time: %v\n", resp.Status, duration)
				lg.mu.Lock()
				lg.countSuccess++
				lg.mu.Unlock()
			}

			// Send metrics to MetricsCollector
			// Example: metricsCollector.Record(duration, resp.StatusCode)
		}
	}

}
