# Prometheus 쿼리 예시 모음

이 문서는 실무에서 자주 사용하는 Prometheus PromQL 쿼리 예시를 정리한 것입니다.

## 목차
- [인프라 메트릭](#인프라-메트릭)
- [Kubernetes 메트릭](#kubernetes-메트릭)
- [애플리케이션 메트릭](#애플리케이션-메트릭)
- [비즈니스 메트릭](#비즈니스-메트릭)
- [SLI/SLO 메트릭](#슬리slo-메트릭)

---

## 인프라 메트릭

### CPU 사용률

```promql
# 전체 CPU 사용률 (%)
100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# CPU 모드별 사용률
sum by (mode) (rate(node_cpu_seconds_total[5m])) * 100

# 특정 노드의 CPU 사용률
100 - (avg(rate(node_cpu_seconds_total{instance="node1",mode="idle"}[5m])) * 100)
```

### 메모리 사용률

```promql
# 메모리 사용률 (%)
(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100

# 사용 가능한 메모리 (GB)
node_memory_MemAvailable_bytes / 1024 / 1024 / 1024

# 스왑 사용률
(node_memory_SwapTotal_bytes - node_memory_SwapFree_bytes) / node_memory_SwapTotal_bytes * 100
```

### 디스크 사용률

```promql
# 디스크 사용률 (%)
(node_filesystem_size_bytes - node_filesystem_free_bytes) / node_filesystem_size_bytes * 100

# 사용 가능한 디스크 공간 (GB)
node_filesystem_avail_bytes{fstype!~"tmpfs|fuse.lxcfs"} / 1024 / 1024 / 1024

# 디스크 I/O 사용률
rate(node_disk_io_time_seconds_total[5m]) * 100
```

### 네트워크

```promql
# 네트워크 수신 트래픽 (bytes/s)
rate(node_network_receive_bytes_total[5m])

# 네트워크 송신 트래픽 (bytes/s)
rate(node_network_transmit_bytes_total[5m])

# 총 네트워크 트래픽 (MB/s)
sum(rate(node_network_receive_bytes_total[5m]) + rate(node_network_transmit_bytes_total[5m])) / 1024 / 1024

# 네트워크 에러율
rate(node_network_receive_errs_total[5m]) + rate(node_network_transmit_errs_total[5m])
```

---

## Kubernetes 메트릭

### Pod 리소스 사용

```promql
# Pod별 CPU 사용량 (코어)
sum(rate(container_cpu_usage_seconds_total{pod!=""}[5m])) by (pod, namespace)

# Pod별 메모리 사용량 (MB)
sum(container_memory_working_set_bytes{pod!=""}) by (pod, namespace) / 1024 / 1024

# Pod CPU 사용률 (요청 대비 %)
sum(rate(container_cpu_usage_seconds_total{pod!=""}[5m])) by (pod, namespace)
/
sum(kube_pod_container_resource_requests{resource="cpu"}) by (pod, namespace) * 100

# Pod 메모리 사용률 (제한 대비 %)
sum(container_memory_working_set_bytes{pod!=""}) by (pod, namespace)
/
sum(kube_pod_container_resource_limits{resource="memory"}) by (pod, namespace) * 100
```

### Pod 상태

```promql
# 네임스페이스별 Pod 수
count(kube_pod_info) by (namespace)

# Pending 상태 Pod
kube_pod_status_phase{phase="Pending"} > 0

# Failed 상태 Pod
kube_pod_status_phase{phase="Failed"} > 0

# 재시작 횟수
sum(kube_pod_container_status_restarts_total) by (pod, namespace)

# 최근 5분간 재시작된 Pod
sum(increase(kube_pod_container_status_restarts_total[5m])) by (pod, namespace) > 0
```

### Deployment 상태

```promql
# Available 레플리카 vs Desired 레플리카
kube_deployment_status_replicas_available
/
kube_deployment_spec_replicas

# 레플리카 불일치 (Deployment가 원하는 상태가 아님)
(kube_deployment_spec_replicas - kube_deployment_status_replicas_available) > 0

# 업데이트 진행 중인 Deployment
kube_deployment_status_replicas_updated < kube_deployment_spec_replicas
```

### 네임스페이스별 리소스

```promql
# 네임스페이스별 총 CPU 사용량
sum(rate(container_cpu_usage_seconds_total{pod!=""}[5m])) by (namespace)

# 네임스페이스별 총 메모리 사용량 (GB)
sum(container_memory_working_set_bytes{pod!=""}) by (namespace) / 1024 / 1024 / 1024

# 네임스페이스별 Pod 수
count(kube_pod_info) by (namespace)
```

---

## 애플리케이션 메트릭

### HTTP 요청

```promql
# 초당 요청 수 (RPS)
sum(rate(http_requests_total[5m]))

# 엔드포인트별 RPS
sum(rate(http_requests_total[5m])) by (endpoint)

# 메서드별 RPS
sum(rate(http_requests_total[5m])) by (method)

# 상태 코드별 RPS
sum(rate(http_requests_total[5m])) by (status)
```

### 응답 시간 (Latency)

```promql
# 평균 응답 시간
rate(http_request_duration_seconds_sum[5m])
/
rate(http_request_duration_seconds_count[5m])

# 50th percentile (p50)
histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))

# 95th percentile (p95)
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# 99th percentile (p99)
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

# 최대 응답 시간
max_over_time(http_request_duration_seconds[5m])
```

### 에러율

```promql
# 전체 에러율 (%)
sum(rate(http_requests_total{status=~"5.."}[5m]))
/
sum(rate(http_requests_total[5m])) * 100

# 4xx 에러율
sum(rate(http_requests_total{status=~"4.."}[5m]))
/
sum(rate(http_requests_total[5m])) * 100

# 엔드포인트별 에러율
sum(rate(http_requests_total{status=~"5.."}[5m])) by (endpoint)
/
sum(rate(http_requests_total[5m])) by (endpoint) * 100
```

### 동시 접속

```promql
# 현재 활성 요청 수
http_requests_active

# 최근 5분간 최대 동시 접속
max_over_time(http_requests_active[5m])

# 활성 사용자 수
active_users
```

---

## 비즈니스 메트릭

### 매출

```promql
# 시간당 매출
sum(increase(business_revenue_total[1h]))

# 일일 매출
sum(increase(business_revenue_total[1d]))

# 통화별 매출
sum(increase(business_revenue_total[1h])) by (currency)

# 평균 주문 금액
sum(increase(business_revenue_total[1h]))
/
sum(increase(business_orders_total{status="success"}[1h]))
```

### 주문

```promql
# 시간당 주문 수
sum(increase(business_orders_total[1h]))

# 주문 성공률 (%)
sum(increase(business_orders_total{status="success"}[5m]))
/
sum(increase(business_orders_total[5m])) * 100

# 제품 타입별 주문 분포
sum(business_orders_total) by (product_type)

# 실패한 주문 수
sum(increase(business_orders_total{status="failed"}[1h]))
```

### 사용자

```promql
# 신규 가입자 (시간당)
sum(increase(new_users_total[1h]))

# 활성 사용자 추이
active_users

# 사용자 증가율 (일일)
(sum(increase(new_users_total[1d])) - sum(increase(new_users_total[1d] offset 1d)))
/
sum(increase(new_users_total[1d] offset 1d)) * 100
```

---

## SLI/SLO 메트릭

### 가용성 (Availability)

```promql
# 가용성 (성공률 %)
sum(rate(http_requests_total{status!~"5.."}[5m]))
/
sum(rate(http_requests_total[5m])) * 100

# 30일 가용성
sum(rate(http_requests_total{status!~"5.."}[30d]))
/
sum(rate(http_requests_total[30d])) * 100

# Error Budget 잔여량 (%)
# 목표: 99.9% 가용성 (0.1% 에러 허용)
0.1 - (
  sum(rate(http_requests_total{status=~"5.."}[30d]))
  /
  sum(rate(http_requests_total[30d])) * 100
)
```

### 지연시간 (Latency)

```promql
# SLI: 95% 요청이 200ms 이내 응답
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) < 0.2

# 목표 달성률
count(histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) < 0.2)
/
count(histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))) * 100
```

### 처리량 (Throughput)

```promql
# 초당 처리된 요청 수
sum(rate(http_requests_total{status!~"5.."}[5m]))

# 목표 RPS 대비 달성률
sum(rate(http_requests_total{status!~"5.."}[5m])) / 1000 * 100
# (목표가 1000 RPS인 경우)
```

---

## 고급 쿼리

### 트렌드 분석

```promql
# CPU 사용률 1주일 추세 (선형 회귀)
predict_linear(node_cpu_usage[1w], 3600*24*7)

# 디스크 공간 소진 예측 (30일 후)
predict_linear(node_filesystem_free_bytes[7d], 3600*24*30)

# 메모리 사용량 증가율 (일일)
deriv(node_memory_MemAvailable_bytes[1d])
```

### 이상 탐지

```promql
# 표준편차를 벗어나는 CPU 사용률
abs(
  node_cpu_usage - avg_over_time(node_cpu_usage[1h])
)
>
stddev_over_time(node_cpu_usage[1h]) * 2

# 최근 5분간 급격한 트래픽 증가 (100% 이상)
(
  sum(rate(http_requests_total[5m]))
  /
  sum(rate(http_requests_total[5m] offset 5m))
) > 2
```

### 비율 계산

```promql
# 캐시 히트율
sum(rate(cache_hits_total[5m]))
/
(sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m]))) * 100

# 데이터베이스 연결 사용률
pg_stat_activity_count / pg_settings_max_connections * 100

# 메모리 캐시 효율
(redis_keyspace_hits_total / (redis_keyspace_hits_total + redis_keyspace_misses_total)) * 100
```

### 집계 및 그룹화

```promql
# Top 5 CPU 사용 Pod
topk(5, sum(rate(container_cpu_usage_seconds_total{pod!=""}[5m])) by (pod))

# Bottom 5 트래픽 엔드포인트
bottomk(5, sum(rate(http_requests_total[5m])) by (endpoint))

# 네임스페이스별 메모리 사용 순위
sort_desc(sum(container_memory_working_set_bytes{pod!=""}) by (namespace))
```

---

## 유용한 함수

### rate vs increase vs irate

```promql
# rate: 초당 평균 증가율 (부드러운 그래프)
rate(http_requests_total[5m])

# increase: 기간 동안 총 증가량
increase(http_requests_total[1h])

# irate: 순간 증가율 (민감한 그래프, 스파이크 감지)
irate(http_requests_total[5m])
```

### 시간 이동 (offset)

```promql
# 1일 전과 비교
http_requests_total - http_requests_total offset 1d

# 1주일 전 대비 증가율
(http_requests_total - http_requests_total offset 1w)
/
http_requests_total offset 1w * 100
```

### 집계 함수

```promql
# 합계
sum(metric) by (label)

# 평균
avg(metric) by (label)

# 최소/최대
min(metric) by (label)
max(metric) by (label)

# 카운트
count(metric) by (label)

# 표준편차
stddev(metric) by (label)
```

---

## 참고 자료

- [PromQL Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [PromQL Functions](https://prometheus.io/docs/prometheus/latest/querying/functions/)
- [PromQL Examples](https://prometheus.io/docs/prometheus/latest/querying/examples/)
