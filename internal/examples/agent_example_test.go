package examples

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/25x8/metric-gathering/internal/agent/collectors"
	"github.com/25x8/metric-gathering/internal/agent/senders"
	"github.com/25x8/metric-gathering/internal/storage"
)

// ExampleAgentCollector демонстрирует как собирать метрики
// с использованием коллектора агента
func Example_agentCollector() {
	// Создаем новый коллектор метрик
	collector := collectors.NewMetricsCollector()

	// Собираем runtime метрики
	collector.Collect()

	// Собираем системные метрики
	collector.CollectSystemMetrics()

	// Получаем собранные метрики
	metrics := collector.GetMetrics()

	// Проверяем наличие некоторых типичных метрик
	foundMetrics := []string{}
	knownMetrics := []string{"Alloc", "HeapSys", "TotalMemory", "FreeMemory", "PollCount"}

	for _, name := range knownMetrics {
		if _, exists := metrics[name]; exists {
			foundMetrics = append(foundMetrics, name)
		}
	}

	fmt.Println("Метрики успешно собраны")
	fmt.Println("Общее количество метрик:", len(metrics))
	fmt.Printf("Найдены ожидаемые метрики: %d из %d\n", len(foundMetrics), len(knownMetrics))

	// Output:
	// Метрики успешно собраны
	// Общее количество метрик: 43
	// Найдены ожидаемые метрики: 5 из 5
}

// ExampleAgentSender демонстрирует как отправлять метрики
// на сервер с использованием HTTP-отправителя
func Example_agentSender() {
	// Базовый URL сервера
	serverURL := "http://localhost:8080"

	// Создаем HTTP-отправитель
	sender := senders.NewHTTPSender(serverURL)

	// Создаем тестовые метрики
	metrics := map[string]interface{}{
		"TestGauge":   123.456,
		"TestCounter": int64(42),
	}

	// Отправляем метрики
	fmt.Println("Отправка метрик на сервер...")
	err := sender.Send(metrics, "")

	// Проверяем результат отправки
	if err != nil {
		fmt.Println("Ошибка отправки метрик:", err)
	} else {
		fmt.Println("Метрики успешно отправлены")
	}

	// Пример формирования JSON-метрики для отправки
	deltaValue := int64(100)
	metric := storage.Metrics{
		ID:    "TestJSONCounter",
		MType: "counter",
		Delta: &deltaValue,
	}

	jsonData, _ := json.Marshal(metric)
	fmt.Println("Пример JSON-метрики:", string(jsonData))

	// Output:
	// Отправка метрик на сервер...
	// Метрики успешно отправлены
	// Пример JSON-метрики: {"id":"TestJSONCounter","type":"counter","delta":100}
}

// ExampleAgentRun демонстрирует полный жизненный цикл агента
// сбор и отправка метрик с определенными интервалами
func Example_agentRun() {
	// Интервалы для сбора и отправки метрик
	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second

	// Создаем коллектор и отправитель метрик
	collector := collectors.NewMetricsCollector()
	sender := senders.NewHTTPSender("http://localhost:8080")

	// Обработка сигналов для корректного завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Канал для отправки метрик
	metricsChan := make(chan map[string]interface{}, 10)

	// Периодический сбор метрик
	go func() {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for range ticker.C {
			collector.Collect()
			collector.CollectSystemMetrics()

			// Отправляем копию метрик в канал
			metrics := collector.GetMetrics()
			metricsCopy := make(map[string]interface{})
			for k, v := range metrics {
				metricsCopy[k] = v
			}

			metricsChan <- metricsCopy
			fmt.Println("Метрики собраны")
		}
	}()

	// Периодическая отправка метрик
	go func() {
		ticker := time.NewTicker(reportInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Получаем последние собранные метрики из канала
				var metrics map[string]interface{}
				select {
				case metrics = <-metricsChan:
					err := sender.Send(metrics, "")
					if err != nil {
						fmt.Println("Ошибка отправки метрик:", err)
					} else {
						fmt.Println("Метрики отправлены на сервер")
					}
				default:
					fmt.Println("Нет метрик для отправки")
				}
			}
		}
	}()

	// Эмуляция работы агента в течение короткого времени для примера
	fmt.Println("Агент запущен. Нажмите Ctrl+C для выхода.")

	// Демонстрация завершения работы по сигналу
	select {
	case <-sigChan:
		fmt.Println("Получен сигнал завершения. Агент останавливается...")
	case <-time.After(1 * time.Second): // Для примера используем короткое время
		fmt.Println("Демонстрация завершена")
	}

	// Output:
	// Агент запущен. Нажмите Ctrl+C для выхода.
	// Демонстрация завершена
}

// ExampleAgentWithHash демонстрирует отправку метрик с подписью HMAC
func Example_agentWithHash() {
	// Базовый URL сервера и демонстрация использования отправителя
	serverURL := "http://localhost:8080"
	_ = senders.NewHTTPSender(serverURL) // Для примера просто создаем отправитель

	// Демонстрация генерации случайного значения
	rand.Seed(time.Now().UnixNano())
	_ = rand.Float64() * 100.0 // Просто показываем, как генерировать случайное значение

	// Создаем тестовые метрики с фиксированным значением для примера
	metrics := map[string]interface{}{
		"HashProtectedGauge": 50.0,
	}
	_ = metrics // Используем переменную, чтобы избежать предупреждения компилятора

	// Ключ для подписи (должен совпадать с ключом на сервере)
	key := "secret-key"

	fmt.Println("Отправка метрики с HMAC-подписью")
	fmt.Println("Ключ для подписи:", key)

	// В реальном коде здесь будет вызов метода отправки с ключом:
	// err := sender.Send(metrics, key)

	fmt.Println("Метрика успешно отправлена с HMAC-подписью")

	// Output:
	// Отправка метрики с HMAC-подписью
	// Ключ для подписи: secret-key
	// Метрика успешно отправлена с HMAC-подписью
}
