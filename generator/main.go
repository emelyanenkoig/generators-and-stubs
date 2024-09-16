package main

import (
	"flag"
	"gns/generator/config"
	loadgen "gns/generator/loadgenerator"
	"gns/generator/metrics"
	"gns/generator/report"
	"log"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	config, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v\n", err)
	}

	metricsCollector := metrics.NewMetricsCollector()
	loadGen := loadgen.NewLoadGenerator(config)

	log.Println("Starting load test...")
	loadGen.Start()

	log.Println("Generating report...")
	report.GenerateReport(metricsCollector)
}
