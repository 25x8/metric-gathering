package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/25x8/metric-gathering/internal/app"
	"github.com/25x8/metric-gathering/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	// Вывод информации о сборке
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	// Определение всех флагов в main
	addrFlag := flag.String("a", "localhost:8080", "HTTP server address")
	storeIntervalFlag := flag.Int("i", 300, "Store interval in seconds (0 for synchronous saving)")
	fileStoragePathFlag := flag.String("f", "/tmp/metrics-db.json", "File storage path")
	restoreFlag := flag.Bool("r", true, "Restore metrics from file at startup")
	databaseDSNFlag := flag.String("d", "", "Database connection string")
	keyFlag := flag.String("k", "", "Secret key for hashing")
	trustedSubnetFlag := flag.String("t", "", "Trusted subnet in CIDR notation")
	memProfile := flag.Bool("memprofile", false, "enable memory profiling")
	flag.Parse()

	// Инициализация логгера и обеспечение его синхронизации
	defer app.SyncLogger()

	h, addr, key, trustedSubnet := app.InitializeAppWithFlags(
		*addrFlag, *storeIntervalFlag, *fileStoragePathFlag,
		*restoreFlag, *databaseDSNFlag, *keyFlag, *trustedSubnetFlag,
	)

	// Получаем конкретную реализацию хранилища для проверки его типа при завершении
	storageImpl := h.Storage

	defer h.CloseDB()

	// Если нужно профилирование, создаем файл профиля
	var profileFile *os.File
	if *memProfile {
		// Создаем директорию profiles, если её нет
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

	r := app.InitializeRouter(h, key, trustedSubnet)

	// Канал для перехвата сигналов завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Server started at %s\n", addr)
		if err := http.ListenAndServe(addr, r); err != nil {
			log.Printf("Server error: %v\n", err)
			// Отправляем сигнал завершения в случае ошибки
			stop <- syscall.SIGTERM
		}
	}()

	// Ожидаем сигнал завершения
	<-stop
	log.Println("Server shutdown initiated...")

	// Если хранилище в памяти, сохраняем его состояние
	if memStorage, ok := storageImpl.(*storage.MemStorage); ok {
		if err := memStorage.Flush(); err != nil {
			log.Printf("Error during flush: %v", err)
		}
	}

	log.Println("Server exited")
}
