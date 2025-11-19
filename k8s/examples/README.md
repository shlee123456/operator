# 예시 파일 가이드

이 디렉토리에는 Prometheus + Grafana + Loki + Alloy 스택을 활용하기 위한 다양한 예시 파일이 포함되어 있습니다.

## 디렉토리 구조

```
examples/
├── exporters/          # 다양한 Exporter 배포 파일
├── alerts/            # AlertManager 및 알림 규칙
├── dashboards/        # Grafana 대시보드 JSON (예정)
└── app-examples/      # 애플리케이션 메트릭 통합 예시
```

## Exporters

### Node Exporter
시스템 메트릭(CPU, 메모리, 디스크 등)을 수집하는 DaemonSet

```bash
kubectl apply -f exporters/node-exporter-daemonset.yaml
```

**수집 메트릭:**
- CPU 사용률
- 메모리 사용률
- 디스크 I/O
- 네트워크 트래픽
- 파일시스템 사용률

**Prometheus 설정 추가:**
```yaml
scrape_configs:
  - job_name: 'node-exporter'
    kubernetes_sd_configs:
    - role: pod
      namespaces:
        names:
        - monitoring
    relabel_configs:
    - source_labels: [__meta_kubernetes_pod_label_app]
      action: keep
      regex: node-exporter
```

### kube-state-metrics
Kubernetes 오브젝트 상태를 메트릭으로 노출

```bash
kubectl apply -f exporters/kube-state-metrics.yaml
```

**수집 메트릭:**
- Pod 상태
- Deployment 상태
- Node 상태
- PVC 상태
- Service 상태

### Blackbox Exporter
HTTP/HTTPS, TCP, ICMP 엔드포인트 모니터링

```bash
kubectl apply -f exporters/blackbox-exporter.yaml
```

**사용 예시:**
```yaml
# Prometheus 설정
scrape_configs:
  - job_name: 'blackbox'
    metrics_path: /probe
    params:
      module: [http_2xx]
    static_configs:
      - targets:
        - https://example.com
        - https://api.example.com
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: blackbox-exporter:9115
```

### Database Exporters

#### PostgreSQL Exporter
```bash
# Secret 수정 (실제 연결 정보로 변경)
vi exporters/postgres-exporter.yaml

kubectl apply -f exporters/postgres-exporter.yaml
```

#### Redis Exporter
```bash
# Secret 수정 (Redis 비밀번호 설정)
vi exporters/redis-exporter.yaml

kubectl apply -f exporters/redis-exporter.yaml
```

## Alerts

### Prometheus Alerts 설정

알림 규칙을 Prometheus에 추가:

```bash
# 1. 알림 규칙 ConfigMap 생성
kubectl apply -f alerts/prometheus-alerts.yaml

# 2. Prometheus ConfigMap에 알림 규칙 파일 추가
kubectl edit configmap prometheus-config -n monitoring
```

Prometheus 설정에 다음 추가:
```yaml
rule_files:
  - '/etc/prometheus/alerts/*.yml'
```

### AlertManager 배포

```bash
# 1. AlertManager 설정 수정 (Slack webhook 등)
vi alerts/alertmanager.yaml

# 2. 배포
kubectl apply -f alerts/alertmanager.yaml

# 3. Prometheus에 AlertManager 연동
kubectl edit configmap prometheus-config -n monitoring
```

Prometheus 설정에 다음 추가:
```yaml
alerting:
  alertmanagers:
  - static_configs:
    - targets:
      - alertmanager:9093
```

### 알림 채널 설정

**Slack 연동:**
1. Slack에서 Incoming Webhook 생성
2. `alerts/alertmanager.yaml`의 `slack_api_url` 수정
3. 채널명 수정 (`#monitoring-alerts`)

**Email 연동:**
1. SMTP 서버 정보 설정
2. `email_configs` 섹션 수정

**PagerDuty 연동:**
1. PagerDuty Service Key 발급
2. `pagerduty_configs` 섹션 수정

## Application Examples

### Go HTTP Server 예시 ⭐ 추천

Go로 작성된 Prometheus 메트릭과 구조화된 로그를 제공하는 HTTP 서버입니다.

```bash
cd app-examples

# 의존성 다운로드
go mod download

# 실행
go run go-http-metrics.go

# 메트릭 확인
curl http://localhost:8080/metrics
```

**상세 가이드:** [go-example-README.md](app-examples/go-example-README.md) 참조

**API 테스트:**
```bash
# 헬스체크
curl http://localhost:8080/health

# 사용자 목록 조회
curl http://localhost:8080/api/users

# 주문 생성 (비즈니스 메트릭 생성)
curl -X POST http://localhost:8080/api/order \
  -H "Content-Type: application/json" \
  -d '{"product_type":"book","amount":29.99,"user_id":"user123"}'

# 느린 엔드포인트
curl http://localhost:8080/api/slow

# 에러 시뮬레이션
curl http://localhost:8080/api/error
```

**Kubernetes 배포:**
```bash
# Docker 이미지 빌드
cd app-examples
docker build -t go-http-metrics:latest .

# Kubernetes 배포
kubectl apply -f go-app-k8s.yaml

# 상태 확인
kubectl get pods -n demo-app

# 포트 포워딩
kubectl port-forward -n demo-app svc/go-http-app 8080:8080
```

### Python Flask 예시

```bash
cd app-examples

# 의존성 설치
pip install flask prometheus-client

# 실행
python python-flask-metrics.py

# 메트릭 확인
curl http://localhost:5000/metrics
```

**API 테스트:**
```bash
# 일반 요청
curl http://localhost:5000/api/users

# 주문 생성
curl -X POST http://localhost:5000/api/order \
  -H "Content-Type: application/json" \
  -d '{"product_type":"book","amount":29.99,"currency":"USD"}'

# 느린 엔드포인트
curl http://localhost:5000/api/slow

# 에러 시뮬레이션
curl http://localhost:5000/api/error
```

### Node.js Express 예시

```bash
cd app-examples

# 의존성 설치
npm install express prom-client

# 실행
node nodejs-express-metrics.js

# 메트릭 확인
curl http://localhost:3000/metrics
```

### Kubernetes 배포

애플리케이션을 Kubernetes에 배포하고 Prometheus로 메트릭 수집:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-app
  namespace: default
  labels:
    app: my-app
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "5000"
    prometheus.io/path: "/metrics"
spec:
  ports:
  - port: 5000
    name: web
  selector:
    app: my-app
```

## Query Examples

### Prometheus 쿼리 예시

`app-examples/prometheus-queries.md` 파일 참조

**주요 쿼리:**
- 인프라 메트릭 (CPU, 메모리, 디스크)
- Kubernetes 메트릭 (Pod, Deployment)
- 애플리케이션 메트릭 (HTTP, 응답시간, 에러율)
- 비즈니스 메트릭 (매출, 주문)
- SLI/SLO 메트릭

### LogQL 쿼리 예시

`app-examples/logql-queries.md` 파일 참조

**주요 쿼리:**
- 로그 필터링 및 검색
- JSON/Logfmt 파싱
- 로그 집계 및 메트릭 생성
- 에러 분석
- 성능 분석

## 대시보드

Grafana에서 대시보드를 생성하거나 공식 대시보드를 임포트할 수 있습니다.

### 추천 대시보드 (Grafana.com)

**인프라:**
- Node Exporter Full: ID 1860
- Kubernetes Cluster Monitoring: ID 7249

**애플리케이션:**
- Prometheus Stats: ID 2
- Loki Dashboard: ID 13639

**데이터베이스:**
- PostgreSQL Database: ID 9628
- Redis Dashboard: ID 11835

### 대시보드 임포트 방법

1. Grafana UI 접속
2. 좌측 메뉴 > Dashboards > Import
3. Dashboard ID 입력 또는 JSON 파일 업로드
4. 데이터소스 선택 (Prometheus, Loki)
5. Import 클릭

## 통합 테스트

모든 컴포넌트가 올바르게 동작하는지 확인:

```bash
# 1. 모든 Pod 상태 확인
kubectl get pods -n monitoring

# 2. Exporter 메트릭 확인
kubectl port-forward -n monitoring svc/node-exporter 9100:9100
curl http://localhost:9100/metrics

# 3. Prometheus 타겟 확인
kubectl port-forward -n monitoring svc/prometheus 9090:9090
# 브라우저에서 http://localhost:9090/targets 접속

# 4. Grafana 데이터소스 확인
kubectl port-forward -n monitoring svc/grafana 3000:3000
# Grafana > Configuration > Data Sources

# 5. AlertManager 확인
kubectl port-forward -n monitoring svc/alertmanager 9093:9093
# 브라우저에서 http://localhost:9093 접속
```

## 문제 해결

### Exporter가 메트릭을 수집하지 않는 경우

```bash
# Pod 로그 확인
kubectl logs -n monitoring deployment/node-exporter

# Prometheus 설정 확인
kubectl get configmap prometheus-config -n monitoring -o yaml

# Prometheus 리로드
kubectl rollout restart deployment/prometheus -n monitoring
```

### AlertManager가 알림을 보내지 않는 경우

```bash
# AlertManager 로그 확인
kubectl logs -n monitoring deployment/alertmanager

# AlertManager 설정 확인
kubectl get configmap alertmanager-config -n monitoring -o yaml

# 알림 테스트
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{"labels":{"alertname":"test","severity":"critical"},"annotations":{"summary":"Test alert"}}]'
```

## 추가 리소스

- [USAGE_GUIDE.md](../USAGE_GUIDE.md) - 활용 방안 전체 가이드
- [README.md](../README.md) - 프로젝트 전체 문서
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Loki Documentation](https://grafana.com/docs/loki/)
