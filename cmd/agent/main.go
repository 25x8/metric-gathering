package main

import (
	"flag"
	"log"
	"time"

	"github.com/25x8/metric-gathering/cmd/agent/collectors"
	"github.com/25x8/metric-gathering/cmd/agent/senders"
)

func main() {
	// Определение флагов
	addr := flag.String("a", "localhost:8080", "HTTP server address")
	reportInterval := flag.Int("r", 10, "Report interval in seconds")
	pollInterval := flag.Int("p", 2, "Poll interval in seconds")

	// Парсинг флагов
	flag.Parse()

	collector := collectors.NewMetricsCollector()
	sender := senders.NewHTTPSender("http://" + *addr)

	tickerPoll := time.NewTicker(time.Duration(*pollInterval) * time.Second)
	tickerReport := time.NewTicker(time.Duration(*reportInterval) * time.Second)

	for {
		select {
		case <-tickerPoll.C:
			metrics := collector.Collect()
			log.Println("Metrics collected:", metrics)

		case <-tickerReport.C:
			metrics := collector.Collect()
			err := sender.Send(metrics)
			if err != nil {
				log.Printf("Error sending metrics: %v", err)
			} else {
				log.Println("Metrics sent successfully")
			}
		}
	}
}
