package loadgen

import (
	"context"
	"fmt"
	c "gns/generator/config"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type LoadGenerator struct {
	config *c.Config
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
}

func (lg *LoadGenerator) sendRequests(ctx context.Context, reqConfig c.RequestConfig) {
	transport := &http.Transport{
		MaxIdleConnsPerHost: 100,
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
	req, err := http.NewRequest(reqConfig.Method, reqConfig.URL, nil) // Add body if needed
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return
	}

	for k, v := range reqConfig.Headers {
		req.Header.Set(k, v)
	}

	var counter atomic.Int32
	var counterSuccess atomic.Int32

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("\nCount of Failed: %v\n", counter.Load())
			fmt.Printf("\nCount of Success: %v\n", counterSuccess.Load())

			return
		default:
			start := time.Now()
			resp, err := client.Do(req)
			duration := time.Since(start)

			if err != nil {
				log.Printf("Request failed: %v\n", err)
				counter.Add(1)
			} else {
				log.Printf("Response Status: %s, Time: %v\n", resp.Status, duration)
				resp.Body.Close()
				counterSuccess.Add(1)
			}

			// Send metrics to MetricsCollector
			// Example: metricsCollector.Record(duration, resp.StatusCode)
		}
	}

}
