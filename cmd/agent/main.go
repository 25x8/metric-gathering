package main

import (
	"context"
	"crypto/rsa"
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
	"github.com/25x8/metric-gathering/internal/buildinfo"
	"github.com/25x8/metric-gathering/internal/config"
	"github.com/25x8/metric-gathering/internal/crypto"
	"github.com/25x8/metric-gathering/internal/utils"
)

// HTTPWorkerConfig содержит конфигурацию для HTTP worker'а
type HTTPWorkerConfig struct {
	MetricsChan <-chan map[string]interface{}
	Sender      senders.HTTPMetricsSender
	WG          *sync.WaitGroup
	Key         string
	PublicKey   *rsa.PublicKey
}

// GRPCWorkerConfig содержит конфигурацию для gRPC worker'а
type GRPCWorkerConfig struct {
	MetricsChan <-chan map[string]interface{}
	Sender      senders.MetricsSender
	WG          *sync.WaitGroup
}

func workerHTTP(ctx context.Context, config HTTPWorkerConfig) {
	defer config.WG.Done()
	for {
		select {
		case <-ctx.Done():
			log.Println("HTTP Worker stopping...")
			return
		case metrics, ok := <-config.MetricsChan:
			if !ok {
				return
			}

			err := config.Sender.SendBatch(metrics, config.PublicKey)
			if err != nil {
				if err.Error() == "data too large for RSA encryption, use individual sends" {
					log.Println("Batch too large for encryption, using individual sends")
				} else {
					log.Printf("Error sending metrics batch: %v, falling back to individual sends", err)
				}

				err = config.Sender.Send(metrics, config.Key, config.PublicKey)
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

func workerGRPC(ctx context.Context, config GRPCWorkerConfig) {
	defer config.WG.Done()
	for {
		select {
		case <-ctx.Done():
			log.Println("gRPC Worker stopping...")
			return
		case metrics, ok := <-config.MetricsChan:
			if !ok {
				return
			}

			err := config.Sender.SendBatch(metrics)
			if err != nil {
				log.Printf("Error sending metrics batch via gRPC: %v", err)
			} else {
				log.Printf("Metrics batch sent successfully via gRPC")
			}
		}
	}
}

func main() {
	buildinfo.PrintBuildInfo()

	addr := flag.String("a", "localhost:8080", "HTTP server address")
	reportInterval := flag.Int("r", 10, "Report interval in seconds")
	pollInterval := flag.Int("p", 2, "Poll interval in seconds")
	keyFlag := flag.String("k", "", "Secret key for hashing")
	rateLimit := flag.Int("l", 2, "Number of outgoing requests")
	memProfile := flag.Bool("memprofile", false, "enable memory profiling")
	cryptoKeyPath := flag.String("crypto-key", "", "Path to public key file for encryption")
	configPath := flag.String("c", "", "Path to JSON config file")
	grpcAddr := flag.String("g", "localhost:3200", "gRPC server address")
	useGRPC := flag.Bool("grpc", false, "Use gRPC instead of HTTP")

	configAltFlag := flag.String("config", "", "Path to JSON config file (alternative)")

	flag.Parse()

	if *configPath == "" && *configAltFlag != "" {
		*configPath = *configAltFlag
	}

	if envConfig := os.Getenv("CONFIG"); envConfig != "" && *configPath == "" {
		*configPath = envConfig
	}

	var cfg *config.AgentConfig
	var err error
	if *configPath != "" {
		cfg, err = config.LoadAgentConfig(*configPath)
		if err != nil {
			log.Printf("Failed to load config from %s: %v. Using defaults and flags.", *configPath, err)
		} else {
			log.Printf("Loaded configuration from %s", *configPath)

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

			if flag.Lookup("g").Value.String() == "localhost:3200" {
				*grpcAddr = cfg.GRPCAddress
			}

			if flag.Lookup("grpc").Value.String() == "false" {
				*useGRPC = cfg.UseGRPC
			}
		}
	}

	var stopProfile func()
	if *memProfile {
		stopProfile = utils.StartMemProfiler("agent_base.pprof")
		defer stopProfile()
	}

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

	if envGRPCAddr := os.Getenv("GRPC_ADDRESS"); envGRPCAddr != "" {
		*grpcAddr = envGRPCAddr
	}

	if envUseGRPC := os.Getenv("USE_GRPC"); envUseGRPC != "" {
		if value, err := strconv.ParseBool(envUseGRPC); err == nil {
			*useGRPC = value
		}
	}

	var publicKey *rsa.PublicKey
	if *cryptoKeyPath != "" && !*useGRPC {
		var err error
		publicKey, err = crypto.LoadPublicKey(*cryptoKeyPath)
		if err != nil {
			log.Fatalf("Failed to load public key: %v", err)
		}
		log.Printf("Public key loaded successfully from %s", *cryptoKeyPath)
	}

	collector := collectors.NewMetricsCollector()

	metricsChan := make(chan map[string]interface{}, 100)
	wg := &sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())

	if *useGRPC {
		log.Printf("Using gRPC to send metrics to %s", *grpcAddr)
		grpcSender, err := senders.NewGRPCSender(*grpcAddr)
		if err != nil {
			log.Fatalf("Failed to create gRPC sender: %v", err)
		}
		defer grpcSender.Close()

		for i := 0; i < *rateLimit; i++ {
			wg.Add(1)
			go workerGRPC(ctx, GRPCWorkerConfig{
				MetricsChan: metricsChan,
				Sender:      grpcSender,
				WG:          wg,
			})
		}
	} else {
		log.Printf("Using HTTP to send metrics to %s", *addr)
		httpSender := senders.NewHTTPSender("http://" + *addr)

		for i := 0; i < *rateLimit; i++ {
			wg.Add(1)
			go workerHTTP(ctx, HTTPWorkerConfig{
				MetricsChan: metricsChan,
				Sender:      httpSender,
				WG:          wg,
				Key:         *keyFlag,
				PublicKey:   publicKey,
			})
		}
	}

	tickerPoll := time.NewTicker(time.Duration(*pollInterval) * time.Second)
	tickerReport := time.NewTicker(time.Duration(*reportInterval) * time.Second)

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
				log.Println("Stopping system metrics collection...")
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
