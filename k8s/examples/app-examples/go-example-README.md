# Go HTTP 서버 with Prometheus 메트릭

Prometheus 메트릭과 구조화된 JSON 로그를 제공하는 Go 마이크로서비스 예제입니다.

## 기능

### Prometheus 메트릭
- `http_requests_total`: HTTP 요청 총 횟수 (method, endpoint, status별)
- `http_request_duration_seconds`: HTTP 요청 처리 시간 히스토그램
- `active_users`: 현재 활성 사용자 수
- `revenue_total_usd`: 총 매출 (USD)
- `orders_completed_total`: 완료된 주문 수 (제품 타입별)
- `errors_total`: 에러 발생 횟수 (타입별)

### API 엔드포인트
- `GET /health`: 헬스체크 엔드포인트
- `GET /api/users`: 사용자 목록 조회
- `POST /api/order`: 주문 생성 (비즈니스 메트릭 생성)
- `GET /api/slow`: 느린 응답 시뮬레이션 (1-3초)
- `GET /api/error`: 에러 시뮬레이션
- `GET /metrics`: Prometheus 메트릭 엔드포인트

### 구조화된 로그
모든 요청과 이벤트를 JSON 형식으로 로깅:
```json
{
  "timestamp": "2025-10-29T10:23:45Z",
  "level": "INFO",
  "message": "HTTP request processed",
  "method": "POST",
  "endpoint": "/api/order",
  "status": 201,
  "response_time_ms": 15.2,
  "user_id": "12345",
  "ip": "192.168.1.100"
}
```

## 로컬 실행

### 1. 의존성 설치
```bash
cd examples/app-examples

# Go 모듈 초기화 및 의존성 다운로드
go mod download
```

### 2. 애플리케이션 실행
```bash
go run go-http-metrics.go
```

서버가 `http://localhost:8080`에서 시작됩니다.

### 3. API 테스트

#### 헬스체크
```bash
curl http://localhost:8080/health
```

#### 사용자 목록 조회
```bash
curl http://localhost:8080/api/users
```

#### 주문 생성 (비즈니스 메트릭 생성)
```bash
curl -X POST http://localhost:8080/api/order \
  -H "Content-Type: application/json" \
  -d '{
    "product_type": "book",
    "amount": 29.99,
    "user_id": "user123"
  }'
```

#### 느린 엔드포인트 테스트
```bash
curl http://localhost:8080/api/slow
```

#### 에러 시뮬레이션
```bash
curl http://localhost:8080/api/error
```

#### Prometheus 메트릭 확인
```bash
curl http://localhost:8080/metrics
```

### 4. 부하 테스트
```bash
# Apache Bench 사용
ab -n 1000 -c 10 http://localhost:8080/api/users

# 또는 hey 사용
hey -n 1000 -c 10 http://localhost:8080/api/users
```

## Docker 빌드 및 실행

### 1. Docker 이미지 빌드
```bash
cd examples/app-examples

docker build -t go-http-metrics:latest .
```

### 2. Docker 컨테이너 실행
```bash
docker run -d -p 8080:8080 --name go-http-app go-http-metrics:latest

# 로그 확인
docker logs -f go-http-app

# 메트릭 확인
curl http://localhost:8080/metrics
```

### 3. 컨테이너 정리
```bash
docker stop go-http-app
docker rm go-http-app
```

## Kubernetes 배포

### 1. 네임스페이스 및 애플리케이션 배포
```bash
kubectl apply -f go-app-k8s.yaml
```

이 명령어는 다음을 생성합니다:
- Namespace: `demo-app`
- Deployment: 2개의 레플리카
- Service (ClusterIP): 클러스터 내부 접근
- Service (NodePort): 외부 접근 (포트 30080)

### 2. 배포 확인
```bash
# Pod 상태 확인
kubectl get pods -n demo-app

# 서비스 확인
kubectl get svc -n demo-app

# 로그 확인
kubectl logs -n demo-app -l app=go-http-app -f
```

### 3. 애플리케이션 접근

#### 포트 포워딩 사용
```bash
kubectl port-forward -n demo-app svc/go-http-app 8080:8080
```

그 다음 `http://localhost:8080`에 접근

#### NodePort 사용
```bash
# Minikube를 사용하는 경우
minikube service go-http-app-external -n demo-app

# 또는 직접 접근
curl http://<NODE_IP>:30080/health
```

### 4. Prometheus 연동

애플리케이션은 이미 Prometheus annotations가 설정되어 있습니다:
```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
```

Prometheus가 자동으로 메트릭을 수집하려면, Prometheus ConfigMap에 다음 설정을 추가하세요:

```yaml
scrape_configs:
  - job_name: 'go-http-app'
    kubernetes_sd_configs:
    - role: pod
      namespaces:
        names:
        - demo-app
    relabel_configs:
    - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
      action: keep
      regex: true
    - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
      action: replace
      target_label: __metrics_path__
      regex: (.+)
    - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
      action: replace
      regex: ([^:]+)(?::\d+)?;(\d+)
      replacement: $1:$2
      target_label: __address__
    - action: labelmap
      regex: __meta_kubernetes_pod_label_(.+)
    - source_labels: [__meta_kubernetes_namespace]
      action: replace
      target_label: kubernetes_namespace
    - source_labels: [__meta_kubernetes_pod_name]
      action: replace
      target_label: kubernetes_pod_name
```

### 5. Grafana 대시보드

Grafana에서 다음 PromQL 쿼리를 사용하여 대시보드를 만들 수 있습니다:

#### 요청률 (RPS)
```promql
sum(rate(http_requests_total{job="go-http-app"}[5m]))
```

#### 평균 응답 시간
```promql
histogram_quantile(0.50, rate(http_request_duration_seconds_bucket{job="go-http-app"}[5m]))
```

#### 95 percentile 응답 시간
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job="go-http-app"}[5m]))
```

#### 에러율
```promql
sum(rate(http_requests_total{job="go-http-app",status=~"5.."}[5m]))
/
sum(rate(http_requests_total{job="go-http-app"}[5m])) * 100
```

#### 엔드포인트별 요청 분포
```promql
sum(rate(http_requests_total{job="go-http-app"}[5m])) by (endpoint)
```

#### 활성 사용자 수
```promql
active_users{job="go-http-app"}
```

#### 시간당 매출
```promql
sum(increase(revenue_total_usd{job="go-http-app"}[1h]))
```

#### 제품 타입별 주문 수
```promql
sum(orders_completed_total{job="go-http-app"}) by (product_type)
```

### 6. Loki 로그 수집

Alloy가 이미 설정되어 있다면, Go 애플리케이션의 JSON 로그가 자동으로 Loki로 전송됩니다.

Grafana Explore에서 LogQL 쿼리 사용:

#### 최근 에러 로그
```logql
{namespace="demo-app", app="go-http-app"} | json | level="ERROR"
```

#### 느린 요청 (응답시간 1초 이상)
```logql
{namespace="demo-app", app="go-http-app"}
| json
| response_time_ms > 1000
```

#### 특정 엔드포인트 로그
```logql
{namespace="demo-app", app="go-http-app"}
| json
| endpoint="/api/order"
```

#### 에러율 계산
```logql
sum(rate({namespace="demo-app", app="go-http-app"} | json | level="ERROR" [5m]))
/
sum(rate({namespace="demo-app", app="go-http-app"} [5m]))
```

## 정리

```bash
# Kubernetes 리소스 삭제
kubectl delete -f go-app-k8s.yaml

# 네임스페이스 삭제 (모든 리소스 포함)
kubectl delete namespace demo-app
```

## 커스터마이징

### 새로운 메트릭 추가

`go-http-metrics.go` 파일에서 메트릭을 추가할 수 있습니다:

```go
// 커스텀 카운터
var customMetric = promauto.NewCounter(
    prometheus.CounterOpts{
        Name: "custom_metric_total",
        Help: "Description of custom metric",
    },
)

// 사용
customMetric.Inc()
```

### 새로운 엔드포인트 추가

```go
func handleCustomEndpoint(w http.ResponseWriter, r *http.Request) {
    // 비즈니스 로직
    json.NewEncoder(w).Encode(map[string]string{
        "status": "success",
    })
}

// main 함수에 추가
http.HandleFunc("/api/custom", metricsMiddleware(handleCustomEndpoint))
```

## 참고 자료

- [Prometheus Go Client](https://github.com/prometheus/client_golang)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/naming/)
- [Go HTTP Server](https://pkg.go.dev/net/http)
- [Structured Logging](https://github.com/sirupsen/logrus)
