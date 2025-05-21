package main

import (
	"context"
	"crypto/rsa"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/25x8/metric-gathering/internal/agent/collectors"
	"github.com/25x8/metric-gathering/internal/agent/senders"
	"github.com/25x8/metric-gathering/internal/config"
	"github.com/25x8/metric-gathering/internal/utils"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func worker(ctx context.Context, metricsChan <-chan map[string]interface{}, sender *senders.HTTPSender, wg *sync.WaitGroup, keyFlag string, publicKey *rsa.PublicKey) {
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

			err := sender.SendBatch(metrics, publicKey)
			if err != nil {
				if err.Error() == "data too large for RSA encryption, use individual sends" {
					log.Println("Batch too large for encryption, using individual sends")
				} else {
					log.Printf("Error sending metrics batch: %v, falling back to individual sends", err)
				}

				err = sender.Send(metrics, keyFlag, publicKey)
				if err != nil {
					log.Printf("Error sending metrics: %v", err)
				} else {
					log.Printf("Metrics sent successfully")
				}
			} else {
				log.Printf("Metrics batch sent successfully")
			}
		}
	}
}

func main() {
	// Вывод информации о сборке
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	// Определение флагов
	addr := flag.String("a", "localhost:8080", "HTTP server address")
	reportInterval := flag.Int("r", 10, "Report interval in seconds")
	pollInterval := flag.Int("p", 2, "Poll interval in seconds")
	keyFlag := flag.String("k", "", "Secret key for hashing")
	rateLimit := flag.Int("l", 2, "Number of outgoing requests")
	memProfile := flag.Bool("memprofile", false, "enable memory profiling")
	cryptoKeyPath := flag.String("crypto-key", "", "Path to public key file for encryption")
	configPath := flag.String("c", "", "Path to JSON config file")

	// Альтернативный флаг для конфига
	configAltFlag := flag.String("config", "", "Path to JSON config file (alternative)")

	// Парсинг флагов
	flag.Parse()

	// Если альтернативный флаг задан, используем его
	if *configPath == "" && *configAltFlag != "" {
		*configPath = *configAltFlag
	}

	// Чтение переменной окружения для пути к конфигу
	if envConfig := os.Getenv("CONFIG"); envConfig != "" && *configPath == "" {
		*configPath = envConfig
	}

	// Загрузка конфигурации
	var cfg *config.AgentConfig
	var err error
	if *configPath != "" {
		cfg, err = config.LoadAgentConfig(*configPath)
		if err != nil {
			log.Printf("Failed to load config from %s: %v. Using defaults and flags.", *configPath, err)
		} else {
			log.Printf("Loaded configuration from %s", *configPath)

			// Применяем значения из конфигурации, только если флаги не установлены
			if flag.Lookup("a").Value.String() == "localhost:8080" {
				*addr = cfg.Address
			}

			if flag.Lookup("r").Value.String() == "10" {
				*reportInterval = cfg.ReportInterval
			}

			if flag.Lookup("p").Value.String() == "2" {
				*pollInterval = cfg.PollInterval
			}

			if flag.Lookup("k").Value.String() == "" {
				*keyFlag = cfg.Key
			}

			if flag.Lookup("l").Value.String() == "2" {
				*rateLimit = cfg.RateLimit
			}

			if flag.Lookup("crypto-key").Value.String() == "" {
				*cryptoKeyPath = cfg.CryptoKey
			}
		}
	}

	// Если нужно профилирование, запускаем его
	var stopProfile func()
	if *memProfile {
		stopProfile = utils.StartMemProfiler("agent_base.pprof")
		defer stopProfile()
	}

	// Чтение переменных окружения (приоритет над конфигурационным файлом)
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

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		*cryptoKeyPath = envCryptoKey
	}

	// Загружаем публичный ключ, если указан путь к файлу
	var publicKey *rsa.PublicKey
	if *cryptoKeyPath != "" {
		var err error
		publicKey, err = utils.LoadPublicKey(*cryptoKeyPath)
		if err != nil {
			log.Fatalf("Failed to load public key: %v", err)
		}
		log.Printf("Public key loaded successfully from %s", *cryptoKeyPath)
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
		go worker(ctx, metricsChan, sender, wg, *keyFlag, publicKey)
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
			case <-ctx.Done():
				log.Println("Stopping metrics collection...")
				return
			case <-tickerPoll.C:
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
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	<-stop

	log.Println("Shutting down gracefully...")

	tickerPoll.Stop()
	tickerReport.Stop()

	cancel()

	close(metricsChan)

	wg.Wait()

	log.Println("Agent stopped")
}
