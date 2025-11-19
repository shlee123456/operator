# Prometheus + Grafana + Loki + Alloy 활용 가이드

이 문서는 모니터링 스택의 다양한 활용 방안과 실제 구현 예시를 제공합니다.

## 목차

1. [인프라 모니터링](#1-인프라-모니터링)
2. [애플리케이션 모니터링](#2-애플리케이션-모니터링)
3. [로그 분석](#3-로그-분석-loki-활용)
4. [비즈니스 메트릭 추적](#4-비즈니스-메트릭-추적)
5. [DevOps/SRE 워크플로우](#5-devopssre-워크플로우)
6. [비용 최적화](#6-비용-최적화)
7. [성능 최적화](#7-성능-최적화)
8. [실제 구현 예시](#8-실제-구현-예시)
9. [알림 설정](#9-알림-설정)
10. [다음 단계](#10-다음-단계)

---

## 1. 인프라 모니터링

### 1.1 Kubernetes 클러스터 모니터링

모든 노드의 시스템 메트릭을 수집하여 클러스터 전체의 건강 상태를 파악합니다.

**모니터링 항목:**
- CPU, 메모리, 디스크, 네트워크 사용률
- 노드별 성능 비교
- 리소스 부족 사전 감지

**Node Exporter 배포:**
```yaml
# examples/node-exporter-daemonset.yaml 참조
```

**활용 시나리오:**
- 클러스터 전체 리소스 사용 현황 대시보드
- 노드별 성능 비교 및 병목 지점 파악
- 특정 노드의 메모리/CPU 부족 시 알림

**Prometheus 쿼리 예시:**
```promql
# 노드별 CPU 사용률
100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# 메모리 사용률
(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100

# 디스크 사용률
(node_filesystem_size_bytes - node_filesystem_free_bytes) / node_filesystem_size_bytes * 100
```

### 1.2 컨테이너/Pod 모니터링

**모니터링 항목:**
- Pod별 CPU/메모리 사용량
- 컨테이너 재시작 횟수
- Pod 스케줄링 문제 감지
- 네임스페이스별 리소스 사용량

**필요한 컴포넌트:**
- kube-state-metrics (Kubernetes 오브젝트 메트릭)
- cAdvisor (컨테이너 메트릭, kubelet에 내장)

**Prometheus 쿼리 예시:**
```promql
# Pod별 CPU 사용량
sum(rate(container_cpu_usage_seconds_total{pod!=""}[5m])) by (pod, namespace)

# 메모리 사용량
sum(container_memory_working_set_bytes{pod!=""}) by (pod, namespace)

# 재시작 횟수
kube_pod_container_status_restarts_total

# Pending 상태 Pod
kube_pod_status_phase{phase="Pending"} > 0
```

---

## 2. 애플리케이션 모니터링

### 2.1 웹 애플리케이션 성능

애플리케이션에 Prometheus 클라이언트 라이브러리를 통합하여 커스텀 메트릭을 수집합니다.

**주요 메트릭:**
- `http_requests_total`: 총 HTTP 요청 수
- `http_request_duration_seconds`: 응답 시간
- `http_request_errors_total`: 에러 발생 횟수
- `active_users`: 동시 접속자 수

**Python (Flask) 예시:**
```python
from prometheus_client import Counter, Histogram, Gauge, start_http_server
from flask import Flask, request
import time

app = Flask(__name__)

# 메트릭 정의
REQUEST_COUNT = Counter(
    'http_requests_total',
    'Total HTTP requests',
    ['method', 'endpoint', 'status']
)

REQUEST_DURATION = Histogram(
    'http_request_duration_seconds',
    'HTTP request latency',
    ['method', 'endpoint']
)

ACTIVE_USERS = Gauge(
    'active_users',
    'Number of active users'
)

@app.before_request
def before_request():
    request.start_time = time.time()

@app.after_request
def after_request(response):
    request_latency = time.time() - request.start_time
    REQUEST_DURATION.labels(
        method=request.method,
        endpoint=request.path
    ).observe(request_latency)

    REQUEST_COUNT.labels(
        method=request.method,
        endpoint=request.path,
        status=response.status_code
    ).inc()

    return response

@app.route('/api/data')
def get_data():
    ACTIVE_USERS.inc()
    try:
        # 비즈니스 로직
        return {"data": "example"}
    finally:
        ACTIVE_USERS.dec()

if __name__ == '__main__':
    # Prometheus 메트릭 서버 시작 (포트 8000)
    start_http_server(8000)
    # Flask 앱 시작
    app.run(host='0.0.0.0', port=5000)
```

**Node.js (Express) 예시:**
```javascript
const express = require('express');
const promClient = require('prom-client');

const app = express();

// Prometheus 기본 메트릭 활성화
promClient.collectDefaultMetrics();

// 커스텀 메트릭
const httpRequestCounter = new promClient.Counter({
  name: 'http_requests_total',
  help: 'Total number of HTTP requests',
  labelNames: ['method', 'route', 'status']
});

const httpRequestDuration = new promClient.Histogram({
  name: 'http_request_duration_seconds',
  help: 'Duration of HTTP requests in seconds',
  labelNames: ['method', 'route']
});

// 미들웨어
app.use((req, res, next) => {
  const start = Date.now();

  res.on('finish', () => {
    const duration = (Date.now() - start) / 1000;

    httpRequestCounter.inc({
      method: req.method,
      route: req.route?.path || req.path,
      status: res.statusCode
    });

    httpRequestDuration.observe({
      method: req.method,
      route: req.route?.path || req.path
    }, duration);
  });

  next();
});

// 메트릭 엔드포인트
app.get('/metrics', async (req, res) => {
  res.set('Content-Type', promClient.register.contentType);
  res.end(await promClient.register.metrics());
});

app.listen(3000);
```

### 2.2 데이터베이스 모니터링

**Exporter 목록:**
- PostgreSQL Exporter: 쿼리 성능, 연결 수, 트랜잭션
- MySQL Exporter: 슬로우 쿼리, 락 대기, 복제 상태
- Redis Exporter: 메모리 사용량, 히트율, 명령어 통계
- MongoDB Exporter: 오퍼레이션 카운터, 연결, 복제

**PostgreSQL Exporter 배포 예시:**
```yaml
# examples/postgres-exporter.yaml 참조
```

**유용한 쿼리:**
```promql
# PostgreSQL 활성 연결 수
pg_stat_activity_count

# 슬로우 쿼리 (1초 이상)
rate(pg_stat_statements_mean_exec_time_seconds{query!~".*pg_stat.*"}[5m]) > 1

# 데이터베이스 크기
pg_database_size_bytes

# 트랜잭션 처리율
rate(pg_stat_database_xact_commit[5m])
```

---

## 3. 로그 분석 (Loki 활용)

### 3.1 애플리케이션 로그 수집

Alloy를 사용하여 애플리케이션 로그를 Loki로 전송하고, Grafana에서 분석합니다.

**로그 분석 시나리오:**
1. 에러 로그 실시간 모니터링
2. 특정 사용자의 행동 추적
3. API 응답 시간 패턴 분석
4. 보안 이벤트 감지
5. 트래픽 패턴 분석

**LogQL 쿼리 예시:**

```logql
# 최근 5분간 ERROR 레벨 로그
{app="backend"} |= "ERROR" [5m]

# 특정 엔드포인트의 로그
{app="api"} | json | endpoint="/api/users"

# 응답 시간이 1초 이상인 요청
{app="api"} | json | response_time > 1000

# 404 에러 발생 빈도 (분당)
sum(rate({app="nginx"} |= "404" [5m]))

# 특정 사용자 ID의 모든 액션
{app="backend"} | json | user_id="12345"

# 에러율 계산
sum(rate({app="api"} |= "error" [5m]))
/
sum(rate({app="api"} [5m]))

# 로그 레벨별 집계
sum(count_over_time({app="backend"} [5m])) by (level)
```

**구조화된 로그 예시:**
```json
{
  "timestamp": "2025-10-29T10:23:45Z",
  "level": "INFO",
  "message": "User login successful",
  "user_id": "12345",
  "endpoint": "/api/auth/login",
  "method": "POST",
  "status": 200,
  "response_time": 150,
  "ip": "192.168.1.100"
}
```

### 3.2 보안 로그 모니터링

**모니터링 항목:**
- 인증 실패 패턴 감지
- 비정상적인 접근 시도 추적
- SSH 로그인 시도 모니터링
- API 남용 감지
- SQL 인젝션 시도 감지

**LogQL 쿼리 예시:**
```logql
# 5분 내 5회 이상 로그인 실패
sum(count_over_time({app="auth"} |= "login failed" [5m])) by (ip) > 5

# SSH 로그인 실패
{job="syslog"} |= "sshd" |= "Failed password"

# API 비정상 호출 (분당 100회 이상)
sum(rate({app="api"} [1m])) by (ip) > 100

# SQL 인젝션 시도 패턴
{app="api"} |~ "(?i)(union|select|drop|delete|insert).*from"
```

---

## 4. 비즈니스 메트릭 추적

### 4.1 실시간 비즈니스 지표

애플리케이션 코드에서 비즈니스 메트릭을 직접 추적합니다.

**커스텀 메트릭 예시:**
```python
from prometheus_client import Counter, Gauge

# 매출
revenue_total = Counter('revenue_total', 'Total revenue', ['currency'])

# 신규 가입자
new_users_total = Counter('new_users_total', 'Total new users')

# 주문 완료
orders_completed = Counter('orders_completed', 'Completed orders', ['product_type'])

# 결제 상태
payment_status = Counter('payment_status_total', 'Payment status', ['status'])

# 장바구니 현재 아이템 수
cart_items = Gauge('cart_items', 'Items in carts')

# 비즈니스 로직에서 사용
def complete_order(order):
    revenue_total.labels(currency='USD').inc(order.amount)
    orders_completed.labels(product_type=order.product_type).inc()
    payment_status.labels(status='success').inc()
```

**Prometheus 쿼리:**
```promql
# 시간당 매출
sum(increase(revenue_total[1h]))

# 신규 가입자 추이
rate(new_users_total[5m]) * 3600

# 결제 성공률
sum(rate(payment_status_total{status="success"}[5m]))
/
sum(rate(payment_status_total[5m])) * 100

# 제품 카테고리별 주문 분포
sum(orders_completed) by (product_type)
```

### 4.2 Grafana 대시보드 활용

**실시간 비즈니스 대시보드 구성:**
- 실시간 매출 현황 (금일/주간/월간)
- 시간대별 트래픽 패턴
- 전환율(Conversion Rate) 추이
- 제품 카테고리별 판매 현황
- 지역별 사용자 분포
- A/B 테스트 결과 비교

---

## 5. DevOps/SRE 워크플로우

### 5.1 CI/CD 파이프라인 모니터링

Jenkins, GitLab CI, GitHub Actions 등의 메트릭을 수집합니다.

**모니터링 항목:**
- 빌드 성공/실패율
- 배포 빈도 (DORA metrics)
- 빌드 소요 시간
- 테스트 커버리지 추이
- 배포 롤백 횟수

**Jenkins Exporter 예시:**
```yaml
# examples/jenkins-exporter.yaml 참조
```

**Prometheus 쿼리:**
```promql
# 빌드 성공률
sum(jenkins_builds_success_total)
/
sum(jenkins_builds_total) * 100

# 평균 빌드 시간
avg(jenkins_build_duration_seconds)

# 일일 배포 빈도
sum(increase(deployments_total[1d]))
```

### 5.2 SLI/SLO 관리

Service Level Indicators와 Service Level Objectives를 정의하고 추적합니다.

**SLI 예시:**
- **가용성**: 99.9% uptime (월간 43분 다운타임 허용)
- **지연시간**: 95% 요청이 200ms 이내 응답
- **에러율**: 전체 요청의 0.1% 미만

**Prometheus 쿼리:**
```promql
# 가용성 계산 (성공률)
sum(rate(http_requests_total{status!~"5.."}[5m]))
/
sum(rate(http_requests_total[5m])) * 100

# 95 percentile 지연시간
histogram_quantile(0.95,
  rate(http_request_duration_seconds_bucket[5m])
)

# 에러율
sum(rate(http_requests_total{status=~"5.."}[5m]))
/
sum(rate(http_requests_total[5m])) * 100

# Error Budget 소진율
1 - (
  sum(rate(http_requests_total{status!~"5.."}[30d]))
  /
  sum(rate(http_requests_total[30d]))
)
```

### 5.3 장애 대응

알림 규칙을 설정하여 문제를 사전에 감지하고 빠르게 대응합니다.

**알림 규칙 예시:**
```yaml
# examples/prometheus-alerts.yaml 참조
```

**알림 레벨:**
- **Warning**: 주의가 필요하지만 즉각적인 조치는 불필요
- **Critical**: 즉시 대응 필요
- **Info**: 정보성 알림

---

## 6. 비용 최적화

### 6.1 리소스 사용 패턴 분석

**활용 방법:**
- 사용률이 낮은 리소스 식별
- 피크 시간대 분석
- 오토스케일링 정책 최적화
- 불필요한 인스턴스 제거
- 리소스 요청/제한 최적화

**Prometheus 쿼리:**
```promql
# CPU 사용률이 10% 미만인 Pod
avg_over_time(
  rate(container_cpu_usage_seconds_total{pod!=""}[24h])
) < 0.1

# 메모리 과다 할당 (할당의 30% 미만 사용)
(container_memory_working_set_bytes / container_spec_memory_limit_bytes) < 0.3

# 디스크 사용률이 낮은 PVC
(kubelet_volume_stats_used_bytes / kubelet_volume_stats_capacity_bytes) < 0.3

# 트래픽이 없는 서비스
sum(rate(http_requests_total[24h])) by (service) == 0
```

### 6.2 비용 추적 대시보드

클라우드 비용과 리소스 사용량을 연결하여 비용 효율성을 분석합니다.

**추적 항목:**
- 네임스페이스별 리소스 사용량
- Pod별 CPU/메모리 비용
- 스토리지 비용
- 네트워크 전송 비용

---

## 7. 성능 최적화

### 7.1 APM (Application Performance Monitoring)

**추적 항목:**
- 데이터베이스 쿼리 응답 시간
- 캐시 히트율
- 외부 API 호출 지연
- 페이지 로드 시간
- 트랜잭션 추적

**Redis 캐시 모니터링:**
```promql
# 캐시 히트율
redis_keyspace_hits_total
/
(redis_keyspace_hits_total + redis_keyspace_misses_total) * 100

# 메모리 사용률
redis_memory_used_bytes / redis_memory_max_bytes * 100

# 연결 수
redis_connected_clients
```

### 7.2 용량 계획 (Capacity Planning)

**분석 내용:**
- 트래픽 증가 추세 예측
- 리소스 확장 시점 결정
- 병목 지점 사전 식별
- 성장에 따른 인프라 요구사항

**Prometheus 쿼리:**
```promql
# 일일 트래픽 증가율
(
  sum(rate(http_requests_total[1d] offset 1d))
  -
  sum(rate(http_requests_total[1d] offset 2d))
)
/
sum(rate(http_requests_total[1d] offset 2d)) * 100

# 리소스 사용 추세 (선형 회귀)
predict_linear(node_memory_MemAvailable_bytes[1w], 3600*24*30)
```

---

## 8. 실제 구현 예시

### 8.1 대시보드 구성

#### Infrastructure Dashboard
**패널 구성:**
1. 클러스터 전체 노드 상태
2. 네임스페이스별 리소스 사용량
3. PVC 사용률
4. 네트워크 트래픽
5. Pod 재시작 이력

#### Application Dashboard
**패널 구성:**
1. HTTP 요청률 (RPS)
2. 응답 시간 분포 (p50, p95, p99)
3. 에러율 추이
4. 활성 사용자 수
5. API 엔드포인트별 성능

#### Business Dashboard
**패널 구성:**
1. 실시간 매출
2. 주문 현황
3. 사용자 행동 패턴
4. 지역별 트래픽
5. 전환율 깔때기

#### Logs Dashboard
**패널 구성:**
1. 로그 레벨별 분포
2. 에러 로그 실시간 피드
3. 특정 키워드 추적
4. 사용자별 액션 로그
5. 응답 시간 히트맵

### 8.2 대시보드 JSON 예시

```
examples/dashboards/ 디렉토리에서 확인:
- infrastructure.json
- application.json
- business.json
- logs.json
```

---

## 9. 알림 설정

### 9.1 AlertManager 연동

Prometheus AlertManager를 설정하여 Slack, Email, PagerDuty 등으로 알림을 전송합니다.

**알림 규칙 예시:**
```yaml
# examples/prometheus-alerts.yaml 참조

주요 알림:
- 노드 다운
- 높은 CPU/메모리 사용률
- Pod 재시작 반복
- 높은 에러율
- 디스크 공간 부족
- 인증서 만료 임박
```

### 9.2 알림 채널 설정

**Slack 연동:**
```yaml
# examples/alertmanager-config.yaml 참조
```

**알림 라우팅:**
- Critical: PagerDuty + Slack
- Warning: Slack
- Info: Email

---

## 10. 다음 단계

### 10.1 추가 컴포넌트 배포

**필수 Exporter:**
```bash
# Node Exporter (시스템 메트릭)
kubectl apply -f examples/node-exporter-daemonset.yaml

# kube-state-metrics (Kubernetes 상태)
kubectl apply -f examples/kube-state-metrics.yaml

# Blackbox Exporter (엔드포인트 헬스체크)
kubectl apply -f examples/blackbox-exporter.yaml
```

### 10.2 AlertManager 설정

```bash
kubectl apply -f examples/alertmanager.yaml
kubectl apply -f examples/prometheus-alerts.yaml
```

### 10.3 대시보드 임포트

Grafana 공식 대시보드:
- Node Exporter Full: ID 1860
- Kubernetes Cluster Monitoring: ID 7249
- Loki Dashboard: ID 13639

### 10.4 장기 보관 설정

Thanos나 Cortex를 사용하여 Prometheus 데이터를 장기 보관합니다.

**Thanos 배포:**
```bash
# examples/thanos/ 디렉토리 참조
```

### 10.5 보안 강화

**권장 사항:**
- Grafana 관리자 비밀번호 변경
- Prometheus 접근 제어 (OAuth2 Proxy)
- TLS/SSL 인증서 설정
- RBAC 강화
- Secret 암호화 (Sealed Secrets)

---

## 참고 자료

### 공식 문서
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Loki Documentation](https://grafana.com/docs/loki/)
- [Alloy Documentation](https://grafana.com/docs/alloy/)

### 커뮤니티
- [PromQL Cheat Sheet](https://promlabs.com/promql-cheat-sheet/)
- [LogQL Cheat Sheet](https://megamorf.gitlab.io/cheat-sheets/loki/)
- [Grafana Dashboards](https://grafana.com/grafana/dashboards/)

### 예시 파일
- `examples/` 디렉토리에 모든 예시 파일 포함
- `dashboards/` 디렉토리에 대시보드 JSON 파일 포함
- `alerts/` 디렉토리에 알림 규칙 파일 포함
