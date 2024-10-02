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
	configPath := flag.String("conf", "conf.json", "Path to configuration file")
	flag.Parse()

	conf, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load conf: %v\n", err)
	}

	metricsCollector := metrics.NewMetricsCollector()
	loadGen := loadgen.NewLoadGenerator(conf)

	log.Println("Starting load test...")
	loadGen.Start()

	log.Println("Generating report...")
	report.GenerateReport(metricsCollector)
}
