//go:build example
// +build example

package examples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/25x8/metric-gathering/internal/storage"
)

// ExampleSendGaugeMetric демонстрирует отправку gauge метрики на сервер
// с использованием URL-параметров.
func Example_sendGaugeMetric() {
	// Подготовка URL с параметрами для отправки метрики
	metricType := "gauge"
	metricName := "ExampleGauge"
	metricValue := "123.456"
	url := fmt.Sprintf("http://localhost:8080/update/%s/%s/%s",
		metricType, metricName, metricValue)

	// Создание и отправка POST-запроса
	response, err := http.Post(url, "text/plain", nil)
	if err != nil {
		fmt.Println("Ошибка отправки запроса:", err)
		return
	}
	defer response.Body.Close()

	// Чтение ответа сервера
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Ошибка чтения ответа:", err)
		return
	}

	fmt.Println("Статус ответа:", response.StatusCode)
	fmt.Println("Тело ответа:", strings.TrimSpace(string(body)))

	// Output:
	// Статус ответа: 200
	// Тело ответа: Metric ExampleGauge updated
}

// ExampleSendCounterMetric демонстрирует отправку counter метрики на сервер
// с использованием URL-параметров.
func Example_sendCounterMetric() {
	// Подготовка URL с параметрами для отправки метрики
	metricType := "counter"
	metricName := "ExampleCounter"
	metricValue := "42"
	url := fmt.Sprintf("http://localhost:8080/update/%s/%s/%s",
		metricType, metricName, metricValue)

	// Создание и отправка POST-запроса
	response, err := http.Post(url, "text/plain", nil)
	if err != nil {
		fmt.Println("Ошибка отправки запроса:", err)
		return
	}
	defer response.Body.Close()

	// Чтение ответа сервера
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Ошибка чтения ответа:", err)
		return
	}

	fmt.Println("Статус ответа:", response.StatusCode)
	fmt.Println("Тело ответа:", strings.TrimSpace(string(body)))

	// Output:
	// Статус ответа: 200
	// Тело ответа: Metric ExampleCounter updated
}

// ExampleSendMetricJSON демонстрирует отправку метрики на сервер
// с использованием JSON формата.
func Example_sendMetricJSON() {
	// Подготовка JSON для отправки метрики
	url := "http://localhost:8080/update/"

	// Создание метрики для gauge типа
	value := 123.456
	metric := storage.Metrics{
		ID:    "ExampleJSONGauge",
		MType: "gauge",
		Value: &value,
	}

	// Сериализация метрики в JSON
	jsonData, err := json.Marshal(metric)
	if err != nil {
		fmt.Println("Ошибка формирования JSON:", err)
		return
	}

	// Создание и отправка POST-запроса с JSON
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Ошибка создания запроса:", err)
		return
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Ошибка отправки запроса:", err)
		return
	}
	defer response.Body.Close()

	// Чтение и десериализация ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Ошибка чтения ответа:", err)
		return
	}

	var respMetric storage.Metrics
	if err := json.Unmarshal(body, &respMetric); err != nil {
		fmt.Println("Ошибка разбора JSON ответа:", err)
		return
	}

	fmt.Println("Статус ответа:", response.StatusCode)
	fmt.Println("ID метрики:", respMetric.ID)
	fmt.Println("Тип метрики:", respMetric.MType)
	if respMetric.Value != nil {
		fmt.Printf("Значение: %.3f\n", *respMetric.Value)
	}

	// Output:
	// Статус ответа: 200
	// ID метрики: ExampleJSONGauge
	// Тип метрики: gauge
	// Значение: 123.456
}

// ExampleGetMetric демонстрирует получение значения метрики
// по имени и типу.
func Example_getMetric() {
	// Подготовка URL для получения метрики
	metricType := "gauge"
	metricName := "ExampleGauge"
	url := fmt.Sprintf("http://localhost:8080/value/%s/%s",
		metricType, metricName)

	// Отправка GET-запроса
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Ошибка отправки запроса:", err)
		return
	}
	defer response.Body.Close()

	// Чтение ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Ошибка чтения ответа:", err)
		return
	}

	fmt.Println("Статус ответа:", response.StatusCode)
	fmt.Println("Значение метрики:", strings.TrimSpace(string(body)))

	// Output:
	// Статус ответа: 200
	// Значение метрики: 123.456
}

// ExampleGetMetricJSON демонстрирует получение значения метрики
// с использованием JSON запроса.
func Example_getMetricJSON() {
	// Подготовка URL и JSON для запроса
	url := "http://localhost:8080/value/"

	metric := storage.Metrics{
		ID:    "ExampleJSONGauge",
		MType: "gauge",
	}

	// Сериализация запроса
	jsonData, err := json.Marshal(metric)
	if err != nil {
		fmt.Println("Ошибка формирования JSON:", err)
		return
	}

	// Создание и отправка POST-запроса
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Ошибка создания запроса:", err)
		return
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Ошибка отправки запроса:", err)
		return
	}
	defer response.Body.Close()

	// Чтение и десериализация ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Ошибка чтения ответа:", err)
		return
	}

	var respMetric storage.Metrics
	if err := json.Unmarshal(body, &respMetric); err != nil {
		fmt.Println("Ошибка разбора JSON ответа:", err)
		return
	}

	fmt.Println("Статус ответа:", response.StatusCode)
	fmt.Println("ID метрики:", respMetric.ID)
	fmt.Println("Тип метрики:", respMetric.MType)
	if respMetric.Value != nil {
		fmt.Printf("Значение: %.3f\n", *respMetric.Value)
	}

	// Output:
	// Статус ответа: 200
	// ID метрики: ExampleJSONGauge
	// Тип метрики: gauge
	// Значение: 123.456
}

// ExampleUpdateBatch демонстрирует пакетное обновление метрик
// с использованием JSON.
func Example_updateBatch() {
	// URL для пакетного обновления
	url := "http://localhost:8080/updates/"

	// Подготовка метрик для пакетного обновления
	gaugeValue := 123.456
	counterValue := int64(42)

	metrics := []storage.Metrics{
		{
			ID:    "BatchGauge",
			MType: "gauge",
			Value: &gaugeValue,
		},
		{
			ID:    "BatchCounter",
			MType: "counter",
			Delta: &counterValue,
		},
	}

	// Сериализация метрик
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		fmt.Println("Ошибка формирования JSON:", err)
		return
	}

	// Создание и отправка POST-запроса
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Ошибка создания запроса:", err)
		return
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Ошибка отправки запроса:", err)
		return
	}
	defer response.Body.Close()

	fmt.Println("Статус ответа:", response.StatusCode)

	// Output:
	// Статус ответа: 200
}
