package app

import (
	"bytes"
	"crypto/rsa"
	"database/sql"
	"encoding/hex"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/25x8/metric-gathering/internal/config"
	"github.com/25x8/metric-gathering/internal/handler"
	"github.com/25x8/metric-gathering/internal/logger"
	"github.com/25x8/metric-gathering/internal/middleware"
	"github.com/25x8/metric-gathering/internal/storage"
	"github.com/25x8/metric-gathering/internal/utils"
	"github.com/gorilla/mux"
)

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

			hashHeader := r.Header.Get("HashSHA256")
			if hashHeader == "" {
				http.Error(w, "Missing HashSHA256 header", http.StatusBadRequest)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			expectedHash := utils.CalculateHash(body, key)
			if hashHeader != expectedHash {
				http.Error(w, "Invalid hash", http.StatusBadRequest)
				return
			}

			w.Header().Set("HashSHA256", expectedHash)

			next.ServeHTTP(w, r)
		})
	}
}

// MiddlewareWithDecryption добавляет расшифровку данных с помощью приватного ключа
func MiddlewareWithDecryption(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privateKey == nil {
				next.ServeHTTP(w, r)
				return
			}

			if r.Header.Get("Content-Encrypted") != "true" {
				next.ServeHTTP(w, r)
				return
			}

			encryptedBody, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading encrypted request body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			decryptedData, err := utils.DecryptWithPrivateKey(encryptedBody, privateKey)
			if err != nil {
				http.Error(w, "Failed to decrypt request data", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decryptedData))
			r.ContentLength = int64(len(decryptedData))

			next.ServeHTTP(w, r)
		})
	}
}

func InitializeApp() (*handler.Handler, string, string, string) {
	addrFlag := flag.String("a", "localhost:8080", "HTTP server address")
	storeIntervalFlag := flag.Int("i", 300, "Store interval in seconds (0 for synchronous saving)")
	fileStoragePathFlag := flag.String("f", "/tmp/metrics-db.json", "File storage path")
	restoreFlag := flag.Bool("r", true, "Restore metrics from file at startup")
	databaseDSNFlag := flag.String("d", "", "Database connection string")
	keyFlag := flag.String("k", "", "Secret key for hashing")
	trustedSubnetFlag := flag.String("t", "", "Trusted subnet in CIDR format")
	configPath := flag.String("c", "", "Path to JSON config file")
	configAltPath := flag.String("config", "", "Path to JSON config file (alternative)")

	flag.Parse()

	if *configPath == "" && *configAltPath != "" {
		*configPath = *configAltPath
	}

	if envConfig := os.Getenv("CONFIG"); envConfig != "" && *configPath == "" {
		*configPath = envConfig
	}

	var cfg *config.ServerConfig
	var err error
	if *configPath != "" {
		cfg, err = config.LoadServerConfig(*configPath)
		if err != nil {
			log.Printf("Failed to load config from %s: %v. Using defaults and flags.", *configPath, err)
		} else {
			log.Printf("Loaded configuration from %s", *configPath)

			if flag.Lookup("a").Value.String() == "localhost:8080" {
				*addrFlag = cfg.Address
			}

			if flag.Lookup("i").Value.String() == "300" {
				*storeIntervalFlag = cfg.StoreInterval
			}

			if flag.Lookup("f").Value.String() == "/tmp/metrics-db.json" {
				*fileStoragePathFlag = cfg.StoreFile
			}

			if flag.Lookup("r").Value.String() == "true" {
				*restoreFlag = cfg.Restore
			}

			if flag.Lookup("d").Value.String() == "" {
				*databaseDSNFlag = cfg.DatabaseDSN
			}

			if flag.Lookup("k").Value.String() == "" {
				*keyFlag = cfg.Key
			}

			if flag.Lookup("t").Value.String() == "" {
				*trustedSubnetFlag = cfg.TrustedSubnet
			}
		}
	}

	// Чтение переменных окружения с приоритетом над конфигурационным файлом
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

	databaseDSN := *databaseDSNFlag
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		databaseDSN = envDatabaseDSN
	}

	key := *keyFlag
	if envKey := os.Getenv("KEY"); envKey != "" {
		key = envKey
	}

	trustedSubnet := *trustedSubnetFlag
	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		trustedSubnet = envTrustedSubnet
	}

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
		if storeInterval > 0 {
			go storage.RunPeriodicSave(memStorage, fileStoragePath, storeInterval)
		}
		log.Println("Using file or in-memory storage")
	}

	h := handler.Handler{
		Storage: storageEngine,
		DB:      dbConnection,
	}

	return &h, addr, key, trustedSubnet
}

func InitializeRouter(h *handler.Handler, key string, privateKeyPath string, trustedSubnet string) *mux.Router {
	r := mux.NewRouter()

	var privateKey *rsa.PrivateKey
	if privateKeyPath != "" {
		var err error
		privateKey, err = utils.LoadPrivateKey(privateKeyPath)
		if err != nil {
			log.Printf("Failed to load private key: %v", err)
		} else {
			log.Printf("Private key loaded successfully from %s", privateKeyPath)
		}
	}

	var wrapWithHash func(http.Handler) http.Handler
	if key != "" {
		wrapWithHash = MiddlewareWithHash(key)
	} else {
		wrapWithHash = func(next http.Handler) http.Handler {
			return next
		}
	}

	var wrapWithDecryption func(http.Handler) http.Handler
	if privateKey != nil {
		wrapWithDecryption = MiddlewareWithDecryption(privateKey)
	} else {
		wrapWithDecryption = func(next http.Handler) http.Handler {
			return next
		}
	}

	wrapWithTrustedSubnet := middleware.TrustedSubnetMiddleware(trustedSubnet)

	wrapHandler := func(handler http.Handler) http.Handler {
		return middleware.GzipMiddleware(
			logger.RequestLogger(
				wrapWithTrustedSubnet(
					wrapWithHash(
						wrapWithDecryption(handler),
					),
				),
			),
		)
	}

	r.Handle("/update/{type}/{name}/{value}", wrapHandler(http.HandlerFunc(h.HandleUpdateMetric))).Methods(http.MethodPost)
	r.Handle("/value/{type}/{name}", wrapHandler(http.HandlerFunc(h.HandleGetValue))).Methods(http.MethodGet)
	r.Handle("/", wrapHandler(http.HandlerFunc(h.HandleGetAllMetrics))).Methods(http.MethodGet)

	r.Handle("/update/", wrapHandler(http.HandlerFunc(h.HandleUpdateMetricJSON))).Methods(http.MethodPost)
	r.Handle("/value/", wrapHandler(http.HandlerFunc(h.HandleGetValueJSON))).Methods(http.MethodPost)

	r.Handle("/ping", wrapHandler(http.HandlerFunc(h.HandlePing))).Methods(http.MethodGet)

	r.Handle("/updates/", wrapHandler(http.HandlerFunc(h.HandleUpdatesBatch))).Methods(http.MethodPost)

	return r
}

func SyncLogger() {
	logger.Sync()
}

func isValidSHA256(key string) bool {
	if len(key) != 64 {
		return false
	}

	_, err := hex.DecodeString(key)
	return err == nil
}
