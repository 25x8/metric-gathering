package main

import (
	"flag"
	"github.com/25x8/metric-gathering/internal/agent/collectors"
	"github.com/25x8/metric-gathering/internal/agent/senders"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	// Определение флагов
	addr := flag.String("a", "localhost:8080", "HTTP server address")
	reportInterval := flag.Int("r", 10, "Report interval in seconds")
	pollInterval := flag.Int("p", 2, "Poll interval in seconds")

	// Парсинг флагов
	flag.Parse()

	// Чтение переменных окружения
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		*addr = envAddr
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		if value, err := strconv.Atoi(envReportInterval); err == nil {
			*reportInterval = value
		}
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		if value, err := strconv.Atoi(envPollInterval); err == nil {
			*pollInterval = value
		}
	}

	collector := collectors.NewMetricsCollector()
	sender := senders.NewHTTPSender("http://" + *addr)

	tickerPoll := time.NewTicker(time.Duration(*pollInterval) * time.Second)
	tickerReport := time.NewTicker(time.Duration(*reportInterval) * time.Second)

	for {
		select {
		case <-tickerPoll.C:
			collector.Collect()
			log.Println("Metrics collected")

		case <-tickerReport.C:
			collector.Collect()
			metrics := collector.GetMetrics()
			if len(metrics) == 0 {
				continue
			}
			err := sender.Send(metrics)

			if err != nil {
				log.Printf("Error sending metrics: %v", err)
			} else {
				log.Println("Metrics sent successfully")
			}
		}
	}
}
