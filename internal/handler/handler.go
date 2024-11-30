package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/25x8/metric-gathering/internal/storage"
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"net/http"
	"strconv"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type Handler struct {
	Storage storage.Storage
	DB      *sql.DB
}

// HandleGetValue - обработчик для получения значения метрики
func (h *Handler) HandleGetValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	metricType := vars["type"]
	metricName := vars["name"]

	switch metricType {
	case Gauge:
		value, err := h.Storage.GetGaugeMetric(metricName)
		if err != nil {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "%v", value)

	case Counter:
		value, err := h.Storage.GetCounterMetric(metricName)
		if err != nil {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "%v", value)

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
	}
}

func (h *Handler) HandleGetValueJSON(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Unsupported content type", http.StatusUnsupportedMediaType)
		return
	}

	var m storage.Metrics
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&m)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if m.ID == "" || m.MType == "" {
		http.Error(w, "ID and MType are required", http.StatusBadRequest)
		return
	}

	switch m.MType {
	case Gauge:
		value, err := h.Storage.GetGaugeMetric(m.ID)
		if err != nil {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		m.Value = &value
	case Counter:
		delta, err := h.Storage.GetCounterMetric(m.ID)
		if err != nil {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		m.Delta = &delta
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// HandleGetAllMetrics - обработчик для получения всех метрик в виде HTML
func (h *Handler) HandleGetAllMetrics(w http.ResponseWriter, r *http.Request) {
	allMetrics := h.Storage.GetAllMetrics()

	w.Header().Set("Content-Type", "text/html")

	tmpl := `
		<html>
		<head><title>Metrics</title></head>
		<body>
			<h1>All Metrics</h1>
			<table border="1">
				<tr>
					<th>Name</th>
					<th>Value</th>
				</tr>
				{{range $name, $value := .}}
				<tr>
					<td>{{$name}}</td>
					<td>{{$value}}</td>
				</tr>
				{{end}}
			</table>
		</body>
		</html>
		`

	t := template.Must(template.New("metrics").Parse(tmpl))
	t.Execute(w, allMetrics)
}

// HandleUpdateMetric - обработчик для обновления метрики
func (h *Handler) HandleUpdateMetric(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	metricType := vars["type"]
	metricName := vars["name"]
	metricValue := vars["value"]

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	switch metricType {
	case Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		h.Storage.SaveGaugeMetric(metricName, value)

	case Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		h.Storage.SaveCounterMetric(metricName, value)

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Metric %s updated", metricName)

}

func (h *Handler) HandleUpdateMetricJSON(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Unsupported content type", http.StatusUnsupportedMediaType)
		return
	}

	var m storage.Metrics
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&m)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if m.ID == "" || m.MType == "" {
		http.Error(w, "ID and MType are required", http.StatusBadRequest)
		return
	}

	switch m.MType {
	case Gauge:
		if m.Value == nil {
			http.Error(w, "Value is required for gauge", http.StatusBadRequest)
			return
		}
		h.Storage.SaveGaugeMetric(m.ID, *m.Value)
		updatedValue, _ := h.Storage.GetGaugeMetric(m.ID)
		m.Value = &updatedValue
	case "counter":
		if m.Delta == nil {
			http.Error(w, "Delta is required for counter", http.StatusBadRequest)
			return
		}
		h.Storage.SaveCounterMetric(m.ID, *m.Delta)
		updatedDelta, _ := h.Storage.GetCounterMetric(m.ID)
		m.Delta = &updatedDelta
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)

}

func (h *Handler) HandlePing(w http.ResponseWriter, r *http.Request) {
	if h.DB == nil {
		http.Error(w, "Database connection is not initialized", http.StatusInternalServerError)
		return
	}

	// Проверка соединения с базой данных
	err := h.DB.PingContext(r.Context())
	if err != nil {
		http.Error(w, "Failed to connect to the database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) CloseDB() {
	if h.DB != nil {
		h.DB.Close()
	}
}

func (h *Handler) HandleUpdatesBatch(w http.ResponseWriter, r *http.Request) {
	var metrics []storage.Metrics

	// Декодирование JSON
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &metrics)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(metrics) == 0 {
		http.Error(w, "Empty metrics batch", http.StatusBadRequest)
		return
	}

	// Обновление метрик в хранилище в рамках одной транзакции
	err = h.Storage.UpdateMetricsBatch(metrics)
	if err != nil {
		http.Error(w, "Failed to update metrics", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
