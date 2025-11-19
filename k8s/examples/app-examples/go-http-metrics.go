package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus 메트릭 정의
var (
	// HTTP 요청 총 횟수
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP 요청 처리 시간
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// 활성 사용자 수
	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of active users",
		},
	)

	// 비즈니스 메트릭: 총 매출
	revenueTotalUSD = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "revenue_total_usd",
			Help: "Total revenue in USD",
		},
	)

	// 비즈니스 메트릭: 주문 완료 수
	ordersCompleted = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_completed_total",
			Help: "Total number of completed orders",
		},
		[]string{"product_type"},
	)

	// 에러 발생 횟수
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type"},
	)
)

// 구조화된 로그 타입
type LogEntry struct {
	Timestamp    string  `json:"timestamp"`
	Level        string  `json:"level"`
	Message      string  `json:"message"`
	Method       string  `json:"method,omitempty"`
	Endpoint     string  `json:"endpoint,omitempty"`
	Status       int     `json:"status,omitempty"`
	ResponseTime float64 `json:"response_time_ms,omitempty"`
	UserID       string  `json:"user_id,omitempty"`
	IP           string  `json:"ip,omitempty"`
}

// 구조화된 로그 출력
func logJSON(entry LogEntry) {
	entry.Timestamp = time.Now().Format(time.RFC3339)
	jsonLog, _ := json.Marshal(entry)
	fmt.Println(string(jsonLog))
}

// 미들웨어: 메트릭 수집 및 로깅
func metricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// ResponseWriter 래퍼 (상태 코드 캡처용)
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// 요청 처리
		next(rw, r)

		// 처리 시간 계산
		duration := time.Since(start).Seconds()
		durationMs := duration * 1000

		// Prometheus 메트릭 업데이트
		httpRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			fmt.Sprintf("%d", rw.statusCode),
		).Inc()

		httpRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Observe(duration)

		// 구조화된 로그 출력
		logEntry := LogEntry{
			Level:        "INFO",
			Message:      "HTTP request processed",
			Method:       r.Method,
			Endpoint:     r.URL.Path,
			Status:       rw.statusCode,
			ResponseTime: durationMs,
			IP:           r.RemoteAddr,
		}

		if rw.statusCode >= 400 {
			logEntry.Level = "ERROR"
		} else if rw.statusCode >= 300 {
			logEntry.Level = "WARN"
		}

		logJSON(logEntry)
	}
}

// ResponseWriter 래퍼
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// API 핸들러들
func handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	// 활성 사용자 시뮬레이션
	activeUsers.Set(float64(rand.Intn(100) + 50))

	users := []map[string]interface{}{
		{"id": "1", "name": "Alice", "email": "alice@example.com"},
		{"id": "2", "name": "Bob", "email": "bob@example.com"},
		{"id": "3", "name": "Charlie", "email": "charlie@example.com"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func handleOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Method not allowed",
		})
		return
	}

	var order struct {
		ProductType string  `json:"product_type"`
		Amount      float64 `json:"amount"`
		UserID      string  `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		errorsTotal.WithLabelValues("invalid_json").Inc()
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid JSON",
		})
		logJSON(LogEntry{
			Level:    "ERROR",
			Message:  "Failed to parse order JSON",
			Endpoint: r.URL.Path,
			UserID:   order.UserID,
		})
		return
	}

	// 비즈니스 메트릭 업데이트
	revenueTotalUSD.Add(order.Amount)
	ordersCompleted.WithLabelValues(order.ProductType).Inc()

	logJSON(LogEntry{
		Level:    "INFO",
		Message:  "Order completed successfully",
		Endpoint: r.URL.Path,
		UserID:   order.UserID,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "success",
		"order_id":     fmt.Sprintf("ORD-%d", rand.Intn(100000)),
		"product_type": order.ProductType,
		"amount":       order.Amount,
	})
}

func handleSlow(w http.ResponseWriter, r *http.Request) {
	// 느린 응답 시뮬레이션 (1-3초)
	delay := time.Duration(rand.Intn(3)+1) * time.Second
	time.Sleep(delay)

	logJSON(LogEntry{
		Level:        "WARN",
		Message:      fmt.Sprintf("Slow endpoint accessed, delay: %v", delay),
		Endpoint:     r.URL.Path,
		ResponseTime: float64(delay.Milliseconds()),
	})

	json.NewEncoder(w).Encode(map[string]string{
		"status": "completed",
		"delay":  delay.String(),
	})
}

func handleError(w http.ResponseWriter, r *http.Request) {
	errorsTotal.WithLabelValues("simulated_error").Inc()

	logJSON(LogEntry{
		Level:    "ERROR",
		Message:  "Simulated error occurred",
		Endpoint: r.URL.Path,
	})

	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "Simulated internal server error",
	})
}

// 백그라운드에서 주기적으로 메트릭 생성 (시뮬레이션용)
func generateBackgroundMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 활성 사용자 수 랜덤 변경
		activeUsers.Set(float64(rand.Intn(150) + 50))

		logJSON(LogEntry{
			Level:   "DEBUG",
			Message: "Background metrics updated",
		})
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	logJSON(LogEntry{
		Level:   "INFO",
		Message: "Starting Go HTTP server with Prometheus metrics",
	})

	// 백그라운드 메트릭 생성 시작
	go generateBackgroundMetrics()

	// 라우트 설정
	http.HandleFunc("/health", metricsMiddleware(handleHealth))
	http.HandleFunc("/api/users", metricsMiddleware(handleUsers))
	http.HandleFunc("/api/order", metricsMiddleware(handleOrder))
	http.HandleFunc("/api/slow", metricsMiddleware(handleSlow))
	http.HandleFunc("/api/error", metricsMiddleware(handleError))

	// Prometheus 메트릭 엔드포인트
	http.Handle("/metrics", promhttp.Handler())

	// 서버 시작
	port := ":8080"
	logJSON(LogEntry{
		Level:   "INFO",
		Message: fmt.Sprintf("Server listening on %s", port),
	})
	logJSON(LogEntry{
		Level:   "INFO",
		Message: "Metrics available at /metrics",
	})

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
