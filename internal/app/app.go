package app

import (
	"database/sql"
	"encoding/hex"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
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

func MiddlewareWithTrustedSubnet(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем IP-адрес из заголовка X-Real-IP
			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				http.Error(w, "Missing X-Real-IP header", http.StatusForbidden)
				return
			}

			// Парсим доверенную подсеть
			_, trustedNet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				http.Error(w, "Invalid trusted subnet configuration", http.StatusInternalServerError)
				return
			}

			// Парсим IP-адрес из заголовка
			clientIP := net.ParseIP(realIP)
			if clientIP == nil {
				http.Error(w, "Invalid IP address in X-Real-IP header", http.StatusBadRequest)
				return
			}

			// Проверяем, входит ли IP в доверенную подсеть
			if !trustedNet.Contains(clientIP) {
				http.Error(w, "IP address not in trusted subnet", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func InitializeAppWithFlags(addr string, storeInterval int, fileStoragePath string, restore bool, databaseDSN string, key string, trustedSubnet string) (*handler.Handler, string, string, string) {
	// Чтение переменных окружения с приоритетом
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr = envAddr
	}

	var storeIntervalDuration time.Duration
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		intervalSec, err := strconv.Atoi(envStoreInterval)
		if err != nil {
			log.Fatalf("Invalid STORE_INTERVAL: %v", err)
		}
		storeIntervalDuration = time.Duration(intervalSec) * time.Second
	} else {
		storeIntervalDuration = time.Duration(storeInterval) * time.Second
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		fileStoragePath = envFileStoragePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		var err error
		restore, err = strconv.ParseBool(envRestore)
		if err != nil {
			log.Fatalf("Invalid RESTORE value: %v", err)
		}
	}

	// Обработка databaseDSN
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		databaseDSN = envDatabaseDSN
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		key = envKey
	}

	// Обработка trusted_subnet
	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		trustedSubnet = envTrustedSubnet
	}

	// Инициализация логгера
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	defer logger.Sync()

	var storageEngine storage.Storage
	var dbConnection *sql.DB
	var memStorage *storage.MemStorage

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
		memStorage = storage.NewMemStorage(fileStoragePath)
		storageEngine = memStorage
		if restore {
			if err := memStorage.Load(); err != nil {
				log.Printf("Error loading metrics from file: %v", err)
			}
		}
		if storeIntervalDuration > 0 {
			go storage.RunPeriodicSave(memStorage, fileStoragePath, storeIntervalDuration)
		}
		log.Println("Using file or in-memory storage")
	}

	// Создаем обработчик с выбранным хранилищем
	h := handler.Handler{
		Storage: storageEngine,
		DB:      dbConnection,
	}

	return &h, addr, key, trustedSubnet
}

func InitializeRouter(h *handler.Handler, key string, trustedSubnet string) *mux.Router {
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

	// Добавляем middleware для проверки доверенной подсети
	var wrapWithTrustedSubnet func(http.Handler) http.Handler
	if trustedSubnet != "" {
		wrapWithTrustedSubnet = MiddlewareWithTrustedSubnet(trustedSubnet)
	} else {
		wrapWithTrustedSubnet = func(next http.Handler) http.Handler {
			return next // Если trusted_subnet не задан, просто возвращаем обработчик
		}
	}

	// Функция для обертки обработчиков
	wrapHandler := func(handler http.Handler) http.Handler {
		return middleware.GzipMiddleware(
			logger.RequestLogger(
				wrapWithTrustedSubnet(
					wrapWithHash(handler),
				),
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
