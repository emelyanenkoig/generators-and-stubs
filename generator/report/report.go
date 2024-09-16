package report

import (
	"encoding/json"
	"fmt"
	m "gns/generator/metrics"
	"os"
)

type Report struct {
	TPS          float64 `json:"tps"`
	AvgLatency   float64 `json:"avg_latency"`
	SuccessRate  float64 `json:"success_rate"`
	FailureCount int     `json:"failure_count"`
}

func GenerateReport(metrics *m.MetricsCollector) {
	report := Report{
		TPS:          calculateTPS(metrics),
		AvgLatency:   calculateAvgLatency(metrics),
		SuccessRate:  calculateSuccessRate(metrics),
		FailureCount: metrics.GetFailures(),
	}

	jsonReport, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("Error generating report: %v\n", err)
		return
	}

	fmt.Println(string(jsonReport))
	os.WriteFile("report.json", jsonReport, 0644)
}

// Placeholder functions for calculations
func calculateTPS(metrics *m.MetricsCollector) float64         { return 0.0 }
func calculateAvgLatency(metrics *m.MetricsCollector) float64  { return 0.0 }
func calculateSuccessRate(metrics *m.MetricsCollector) float64 { return 0.0 }
