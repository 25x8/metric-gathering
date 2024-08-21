package main

import (
	"log"
	"time"

	"github.com/25x8/metric-gathering/cmd/agent/collectors"
	"github.com/25x8/metric-gathering/cmd/agent/senders"
)

func main() {
	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second

	collector := collectors.NewMetricsCollector()
	sender := senders.NewHTTPSender("http://localhost:8080")

	tickerPoll := time.NewTicker(pollInterval)
	tickerReport := time.NewTicker(reportInterval)

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
