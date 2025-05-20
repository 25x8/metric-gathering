//go:build example
// +build example

package examples

import (
	"context"
	"log"
	"time"

	"github.com/25x8/metric-gathering/internal/agent/collectors"
	"github.com/25x8/metric-gathering/internal/agent/senders"
)

// Example_agentCollector демонстрирует как собирать метрики
// с использованием коллектора агента
func Example_agentCollector() {
	// Создаем новый коллектор метрик
	collector := collectors.NewMetricsCollector()

	// Собираем метрики
	collector.Collect()

	// Получаем все собранные метрики
	metrics := collector.GetMetrics()

	// Выводим некоторые метрики
	log.Println("Собранные метрики:")
	for key, value := range metrics {
		log.Printf("%s: %v\n", key, value)
		// Выводим только первые 3 метрики для краткости примера
		if len(metrics) > 3 {
			log.Println("...")
			break
		}
	}

	// Вывод тестового примера
	// Output:
	// Собранные метрики:
}

// Example_agentSender демонстрирует как отправлять метрики
// на сервер с использованием HTTP-отправителя
func Example_agentSender() {
	// Базовый URL сервера
	serverURL := "http://localhost:8080"

	// Создаем отправителя
	sender := senders.NewHTTPSender(serverURL)

	// Подготавливаем метрики для отправки
	metrics := map[string]interface{}{
		"TestCounter": int64(42),
		"TestGauge":   float64(123.456),
	}

	// Отправляем метрики
	log.Println("Отправка метрик на сервер...")
	err := sender.SendBatch(metrics, nil)
	if err != nil {
		log.Printf("Ошибка отправки метрик: %v\n", err)
	} else {
		log.Println("Метрики успешно отправлены")
	}

	// Пример использования JSON метрик
	log.Println("Пример JSON-метрики: {\"id\":\"TestJSONCounter\",\"type\":\"counter\",\"delta\":100}")

	// Output:
	// Отправка метрик на сервер...
	// Метрики успешно отправлены
	// Пример JSON-метрики: {"id":"TestJSONCounter","type":"counter","delta":100}
}

// Example_agentRun демонстрирует полный жизненный цикл агента
// сбор и отправка метрик с определенными интервалами
func Example_agentRun() {
	// Интервалы для сбора и отправки метрик
	pollInterval := 2 * time.Second
	reportInterval := 5 * time.Second
	maxRuns := 2

	// Создаем контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем коллектор и отправитель
	collector := collectors.NewMetricsCollector()
	sender := senders.NewHTTPSender("http://localhost:8080")

	// Запускаем горутину для сбора метрик
	go func() {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		runs := 0
		for {
			select {
			case <-ticker.C:
				collector.Collect()
				log.Println("Метрики собраны")
				runs++
				if runs >= maxRuns {
					cancel() // Отменяем контекст после заданного числа сборов
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Запускаем горутину для отправки метрик
	go func() {
		ticker := time.NewTicker(reportInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metrics := collector.GetMetrics()
				err := sender.SendBatch(metrics, nil)
				if err != nil {
					log.Printf("Ошибка отправки метрик: %v\n", err)
				} else {
					log.Println("Метрики успешно отправлены")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Ждем завершения (в реальном коде здесь могла бы быть обработка сигналов)
	<-ctx.Done()
	log.Println("Агент завершил работу")

	// Output:
	// Метрики собраны
	// Метрики собраны
	// Агент завершил работу
}

// Example_agentWithHash демонстрирует отправку метрик с подписью HMAC
func Example_agentWithHash() {
	// Базовый URL сервера и демонстрация использования отправителя
	serverURL := "http://localhost:8080"
	key := "secret" // Ключ для подписи

	// Создаем отправителя
	sender := senders.NewHTTPSender(serverURL)

	// Подготавливаем метрики для отправки
	metrics := map[string]interface{}{
		"TestCounter": int64(42),
		"TestGauge":   float64(123.456),
	}

	// Отправляем метрики с подписью
	log.Println("Отправка метрик с подписью...")
	err := sender.Send(metrics, key, nil)
	if err != nil {
		log.Printf("Ошибка отправки метрик: %v\n", err)
	} else {
		log.Println("Метрики успешно отправлены с использованием HMAC подписи")
	}

	// Output:
	// Отправка метрик с подписью...
	// Метрики успешно отправлены с использованием HMAC подписи
}
