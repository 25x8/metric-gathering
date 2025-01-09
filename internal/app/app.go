package app

import (
	"database/sql"
	"encoding/hex"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/25x8/metric-gathering/internal/handler"
	"github.com/25x8/metric-gathering/internal/logger"
	"github.com/25x8/metric-gathering/internal/middleware"
	"github.com/25x8/metric-gathering/internal/storage"
	"github.com/25x8/metric-gathering/internal/utils"
	"github.com/gorilla/mux"
)

// добавляем проверку хеша входящего запроса и генерацию хеша ответа
func MiddlewareWithHash(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !isValidSHA256(key) {
				next.ServeHTTP(w, r)
				return
			}

			// Проверяем заголовок HashSHA256
			hashHeader := r.Header.Get("HashSHA256")
			if hashHeader == "" {
				http.Error(w, "Missing HashSHA256 header", http.StatusBadRequest)
				return
			}

			// Считываем тело запроса
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			// Вычисляем ожидаемый хеш
			expectedHash := utils.CalculateHash(body, key)
			if hashHeader != expectedHash {
				http.Error(w, "Invalid hash", http.StatusBadRequest)
				return
			}

			// Устанавливаем заголовок ответа с хешем
			w.Header().Set("HashSHA256", expectedHash)

			// Передаём управление следующему обработчику
			next.ServeHTTP(w, r)
		})
	}
}

func InitializeApp() (*handler.Handler, string, string) {
	// Определение флагов
	addrFlag := flag.String("a", "localhost:8080", "HTTP server address")
	storeIntervalFlag := flag.Int("i", 300, "Store interval in seconds (0 for synchronous saving)")
	fileStoragePathFlag := flag.String("f", "/tmp/metrics-db.json", "File storage path")
	restoreFlag := flag.Bool("r", true, "Restore metrics from file at startup")
	databaseDSNFlag := flag.String("d", "", "Database connection string")
	keyFlag := flag.String("k", "", "Secret key for hashing")

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

	key := *keyFlag
	if envKey := os.Getenv("KEY"); envKey != "" {
		key = envKey
	}

	// Инициализация логгера
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	defer logger.Sync()

	var storageEngine storage.Storage
	var dbConnection *sql.DB

	// Выбор хранилища
	if databaseDSN != "" {
		db, errOpenDB := sql.Open("pgx", databaseDSN)

		if errOpenDB != nil {
			log.Fatalf("failed to open database: %v", errOpenDB)
		}

		dbStorage, err := storage.NewDBStorage(db)
		if err != nil {
			log.Fatalf("Failed to initialize database storage: %v", err)
		}
		storageEngine = dbStorage
		dbConnection = dbStorage.DB()
		log.Println("Using PostgreSQL storage")

	} else {
		memStorage := storage.NewMemStorage(fileStoragePath)
		storageEngine = memStorage
		if restore {
			if err := memStorage.Load(); err != nil {
				log.Printf("Error loading metrics from file: %v", err)
			}
		}
		if storeInterval > 0 {
			go storage.RunPeriodicSave(memStorage, fileStoragePath, storeInterval)
		}
		log.Println("Using file or in-memory storage")
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			if err := memStorage.Flush(); err != nil {
				log.Printf("Error during flush: %v", err)
			}
			os.Exit(0)
		}()
	}

	// Создаем обработчик с выбранным хранилищем
	h := handler.Handler{
		Storage: storageEngine,
		DB:      dbConnection,
	}

	return &h, addr, key
}

func InitializeRouter(h *handler.Handler, key string) *mux.Router {
	r := mux.NewRouter()

	// Проверяем, передан ли key, и определяем middleware
	var wrapWithHash func(http.Handler) http.Handler
	if key != "" {
		wrapWithHash = MiddlewareWithHash(key)
	} else {
		wrapWithHash = func(next http.Handler) http.Handler {
			return next // Если key не задан, просто возвращаем обработчик
		}
	}

	// Функция для обертки обработчиков
	wrapHandler := func(handler http.Handler) http.Handler {
		return middleware.GzipMiddleware(
			logger.RequestLogger(
				wrapWithHash(handler),
			),
		)
	}

	// Маршруты для обновления метрик и получения их значений
	r.Handle("/update/{type}/{name}/{value}", wrapHandler(http.HandlerFunc(h.HandleUpdateMetric))).Methods(http.MethodPost)
	r.Handle("/value/{type}/{name}", wrapHandler(http.HandlerFunc(h.HandleGetValue))).Methods(http.MethodGet)
	r.Handle("/", wrapHandler(http.HandlerFunc(h.HandleGetAllMetrics))).Methods(http.MethodGet)

	// Маршруты для работы с JSON
	r.Handle("/update/", wrapHandler(http.HandlerFunc(h.HandleUpdateMetricJSON))).Methods(http.MethodPost)
	r.Handle("/value/", wrapHandler(http.HandlerFunc(h.HandleGetValueJSON))).Methods(http.MethodPost)

	// Добавляем маршрут для /ping
	r.Handle("/ping", wrapHandler(http.HandlerFunc(h.HandlePing))).Methods(http.MethodGet)

	// Добавляем маршрут для /updates/
	r.Handle("/updates/", wrapHandler(http.HandlerFunc(h.HandleUpdatesBatch))).Methods(http.MethodPost)

	return r
}

func SyncLogger() {
	logger.Sync()
}

func isValidSHA256(key string) bool {
	// SHA-256 хэш всегда длиной 64 символа
	if len(key) != 64 {
		return false
	}

	// Проверяем, что строка состоит из шестнадцатеричных символов
	_, err := hex.DecodeString(key)
	return err == nil
}
