package app

import (
	"database/sql"
	"flag"
	"github.com/25x8/metric-gathering/internal/handler"
	"github.com/25x8/metric-gathering/internal/logger"
	"github.com/25x8/metric-gathering/internal/middleware"
	"github.com/25x8/metric-gathering/internal/storage"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func InitializeApp() (*handler.Handler, string) {
	// Определение флагов
	addrFlag := flag.String("a", "localhost:8080", "HTTP server address")
	storeIntervalFlag := flag.Int("i", 300, "Store interval in seconds (0 for synchronous saving)")
	fileStoragePathFlag := flag.String("f", "/tmp/metrics-db.json", "File storage path")
	restoreFlag := flag.Bool("r", true, "Restore metrics from file at startup")
	databaseDSNFlag := flag.String("d", "", "Database connection string")

	// Парсинг флагов
	flag.Parse()

	// Чтение переменных окружения с приоритетом
	addr := *addrFlag
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr = envAddr
	}

	var storeInterval time.Duration
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		intervalSec, err := strconv.Atoi(envStoreInterval)
		if err != nil {
			log.Fatalf("Invalid STORE_INTERVAL: %v", err)
		}
		storeInterval = time.Duration(intervalSec) * time.Second
	} else if storeIntervalFlag != nil {
		storeInterval = time.Duration(*storeIntervalFlag) * time.Second
	} else {
		storeInterval = 300 * time.Second
	}

	fileStoragePath := *fileStoragePathFlag
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		fileStoragePath = envFileStoragePath
	}

	restore := *restoreFlag
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		var err error
		restore, err = strconv.ParseBool(envRestore)
		if err != nil {
			log.Fatalf("Invalid RESTORE value: %v", err)
		}
	}

	// Обработка databaseDSN
	databaseDSN := *databaseDSNFlag
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		databaseDSN = envDatabaseDSN
	}

	// Инициализация логгера
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Создание экземпляра MemStorage
	newStorage := storage.NewMemStorage()

	// Восстановление метрик из файла при старте
	if restore {
		if err := storage.LoadFromFile(newStorage, fileStoragePath); err != nil {
			log.Printf("Error loading metrics from file: %v", err)
		}
	}

	// Запуск периодического сохранения метрик
	if storeInterval > 0 {
		go storage.RunPeriodicSave(newStorage, fileStoragePath, storeInterval)
	}

	// Обработка сигнала завершения для сохранения метрик
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if err := storage.SaveToFile(newStorage, fileStoragePath); err != nil {
			log.Printf("Error saving metrics on shutdown: %v", err)
		}
		os.Exit(0)
	}()

	var dbConnection *sql.DB

	if databaseDSN != "" {
		var err error
		dbConnection, err = sql.Open("pgx", databaseDSN)
		if err != nil {
			log.Fatalf("Unable to open database connection: %v", err)
		}

		// Проверка соединения
		if err = dbConnection.Ping(); err != nil {
			log.Fatalf("Unable to connect to database: %v", err)
		}

		log.Println("Connected to the database successfully")
	} else {
		log.Println("Database DSN is not provided")
	}

	// Создаем обработчик с хранилищем
	h := handler.Handler{
		Storage: newStorage,
		DB:      dbConnection,
	}

	return &h, addr
}

func InitializeRouter(h *handler.Handler) *mux.Router {
	r := mux.NewRouter()

	// Маршруты для обновления метрик и получения их значений
	r.Handle("/update/{type}/{name}/{value}", middleware.GzipMiddleware(logger.RequestLogger(http.HandlerFunc(h.HandleUpdateMetric)))).Methods(http.MethodPost)
	r.Handle("/value/{type}/{name}", middleware.GzipMiddleware(logger.RequestLogger(http.HandlerFunc(h.HandleGetValue)))).Methods(http.MethodGet)
	r.Handle("/", middleware.GzipMiddleware(logger.RequestLogger(http.HandlerFunc(h.HandleGetAllMetrics)))).Methods(http.MethodGet)

	// Маршруты для работы с JSON
	r.Handle("/update/", middleware.GzipMiddleware(logger.RequestLogger(http.HandlerFunc(h.HandleUpdateMetricJSON)))).Methods(http.MethodPost)
	r.Handle("/value/", middleware.GzipMiddleware(logger.RequestLogger(http.HandlerFunc(h.HandleGetValueJSON)))).Methods(http.MethodPost)

	// Добавляем маршрут для /ping
	r.Handle("/ping", middleware.GzipMiddleware(logger.RequestLogger(http.HandlerFunc(h.HandlePing)))).Methods(http.MethodGet)

	return r
}

func SyncLogger() {
	logger.Sync()
}