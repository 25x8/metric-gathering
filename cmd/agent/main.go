package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/25x8/metric-gathering/internal/agent/collectors"
	"github.com/25x8/metric-gathering/internal/agent/senders"
)

func worker(ctx context.Context, metricsChan <-chan map[string]interface{}, sender *senders.HTTPSender, wg *sync.WaitGroup, keyFlag string) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopping...")
			return
		case metrics, ok := <-metricsChan:
			if !ok {
				return
			}

			err := sender.Send(metrics, keyFlag)

			if err != nil {
				log.Printf("Error sending metrics: %v", err)
			} else {
				log.Printf("Metrics sent successfully")
			}
		}
	}
}

func main() {
	// Определение флагов
	addr := flag.String("a", "localhost:8080", "HTTP server address")
	reportInterval := flag.Int("r", 10, "Report interval in seconds")
	pollInterval := flag.Int("p", 2, "Poll interval in seconds")
	keyFlag := flag.String("k", "", "Secret key for hashing")
	rateLimit := flag.Int("l", 2, "Number of outgoing requests")

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

	if envKey := os.Getenv("KEY"); envKey != "" {
		*keyFlag = envKey
	}

	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		if value, err := strconv.Atoi(envRateLimit); err == nil {
			*rateLimit = value
		}
	}

	collector := collectors.NewMetricsCollector()
	sender := senders.NewHTTPSender("http://" + *addr)

	tickerPoll := time.NewTicker(time.Duration(*pollInterval) * time.Second)
	tickerReport := time.NewTicker(time.Duration(*reportInterval) * time.Second)

	metricsChan := make(chan map[string]interface{}, 100)
	wg := &sync.WaitGroup{}

	// Контекст для gracefull shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Worker pool
	for i := 0; i < *rateLimit; i++ {
		wg.Add(1)
		go worker(ctx, metricsChan, sender, wg, *keyFlag)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping metrics collection...")
				return
			case <-tickerPoll.C:
				collector.Collect()
				log.Println("Metrics collected")
			}
		}
	}()

	go func() {
		for {
			select {
			case <- ctx.Done():
				log.Println("Stopping metrics collection...")
				return
			case <- tickerPoll.C:
				collector.CollectSystemMetrics()
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping metrics reporting...")
				return
			case <-tickerReport.C:
				metrics := collector.GetMetrics()
				if len(metrics) == 0 {
					continue
				}

				metricsChan <- metrics
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	log.Println("Shutting down gracefully...")

	tickerPoll.Stop()
	tickerReport.Stop()

	cancel()

	close(metricsChan)

	wg.Wait()

	log.Println("Agent stopped")
}
