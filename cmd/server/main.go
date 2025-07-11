package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/25x8/metric-gathering/internal/app"
	"github.com/25x8/metric-gathering/internal/buildinfo"
	"github.com/25x8/metric-gathering/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func init() {
	flag.String("a", "localhost:8080", "HTTP server address")
	flag.Int("i", 300, "Store interval in seconds (0 for synchronous saving)")
	flag.String("f", "/tmp/metrics-db.json", "File storage path")
	flag.Bool("r", true, "Restore metrics from file at startup")
	flag.String("d", "", "Database connection string")
	flag.String("k", "", "Secret key for hashing")
}

func main() {
	buildinfo.PrintBuildInfo()

	memProfile := flag.Bool("memprofile", false, "enable memory profiling")
	cryptoKeyPath := flag.String("crypto-key", "", "Path to private key file for decryption")
	configPath := flag.String("c", "", "Path to JSON config file")
	configAltPath := flag.String("config", "", "Path to JSON config file (alternative)")

	flag.Parse()

	if *configPath == "" && *configAltPath != "" {
		*configPath = *configAltPath
	}

	if envConfig := os.Getenv("CONFIG"); envConfig != "" && *configPath == "" {
		*configPath = envConfig
	}

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		*cryptoKeyPath = envCryptoKey
	}

	defer app.SyncLogger()

	h, addr, key := app.InitializeApp()

	privateKeyPath := *cryptoKeyPath

	storageImpl := h.Storage

	defer h.CloseDB()

	var profileFile *os.File
	if *memProfile {
		if err := os.MkdirAll("profiles", 0755); err != nil {
			log.Fatalf("Failed to create profiles directory: %v", err)
		}

		var err error
		profileFile, err = os.Create("profiles/base.pprof")
		if err != nil {
			log.Fatalf("Failed to create profile file: %v", err)
		}
		defer func() {
			// Записываем профиль перед закрытием файла
			log.Println("Writing memory profile...")
			if err := pprof.WriteHeapProfile(profileFile); err != nil {
				log.Printf("Failed to write profile: %v", err)
			}
			if err := profileFile.Close(); err != nil {
				log.Printf("Failed to close profile file: %v", err)
			}
			log.Println("Profile saved to profiles/base.pprof")
		}()
	}

	r := app.InitializeRouter(h, key, privateKeyPath)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		log.Printf("Server started at %s\n", addr)
		if err := http.ListenAndServe(addr, r); err != nil {
			log.Printf("Server error: %v\n", err)
			stop <- syscall.SIGTERM
		}
	}()

	<-stop
	log.Println("Server shutdown initiated...")

	if memStorage, ok := storageImpl.(*storage.MemStorage); ok {
		if err := memStorage.Flush(); err != nil {
			log.Printf("Error during flush: %v", err)
		}
	}

	log.Println("Server exited")
}
