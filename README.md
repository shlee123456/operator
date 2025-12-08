# Prometheus Monitoring Stack

Prometheus, Grafana, Loki, Alloy를 활용한 통합 모니터링 스택입니다. Docker와 Kubernetes 두 가지 배포 방식을 지원합니다.

## 프로젝트 구조

```
operator/
├── docker/          # Docker Compose 기반 배포
├── k8s/             # Kubernetes 기반 배포
├── docs/            # 공통 문서
└── README.md        # 이 파일
```

## 구성 요소

- **Prometheus**: 메트릭 수집 및 저장
- **Grafana**: 시각화 대시보드
- **Loki**: 로그 수집 및 저장
- **Alloy**: Grafana Agent (메트릭 및 로그 수집)

## 배포 옵션

### 1. Docker Compose (개발/테스트 환경)

로컬 개발 환경이나 간단한 테스트 환경에 적합합니다.

```bash
cd docker
docker-compose up -d
```

자세한 내용은 [docker/README.md](docker/README.md)를 참고하세요.

**장점:**
- 빠른 설정 및 시작
- 로컬 개발에 적합
- 리소스 사용량이 적음

**사용 사례:**
- 로컬 개발 환경
- POC (Proof of Concept)
- 간단한 테스트 환경

### 2. Kubernetes (프로덕션 환경)

프로덕션 환경이나 대규모 클러스터에 적합합니다.

```bash
cd k8s
kubectl apply -k .
```

**장점:**
- 고가용성 및 확장성
- 자동 복구 및 스케일링
- 프로덕션 환경에 적합

**사용 사례:**
- 프로덕션 환경
- 대규모 인프라 모니터링
- 엔터프라이즈 환경

## 빠른 시작

### Docker로 시작하기

```bash
# 저장소 클론
git clone git@github.com:shlee123456/operator.git
cd operator/docker

# 서비스 시작
docker-compose up -d

# 접속 정보
# - Grafana: http://localhost:3000 (admin/admin)
# - Prometheus: http://localhost:9090
# - Loki: http://localhost:3100
```

### Kubernetes로 시작하기

```bash
# 저장소 클론
git clone git@github.com:shlee123456/operator.git
cd operator/k8s

# 배포
kubectl apply -k .

# 서비스 확인
kubectl get pods -n monitoring

# Grafana 접속
kubectl port-forward -n monitoring svc/grafana 3000:3000
# http://localhost:3000 (admin/admin)
```

## 설정 가이드

### 메트릭 수집 대상 추가

#### Docker 환경
`docker/prometheus/prometheus.yml` 파일을 수정합니다:

```yaml
scrape_configs:
  - job_name: 'node_exporter'
    static_configs:
      - targets: ['10.0.1.10:9100']  # 실제 서버 IP로 변경
        labels:
          instance: 'server-1'
```

#### Kubernetes 환경
`k8s/prometheus/configmap.yaml` 파일을 수정합니다:

```yaml
scrape_configs:
  - job_name: 'node_exporter'
    static_configs:
      - targets: ['10.0.1.10:9100']  # 실제 서버 IP로 변경
        labels:
          instance: 'server-1'
```

### Grafana 대시보드

기본 제공 대시보드:
- Node Exporter Full
- Docker Container Metrics
- Loki Logs

커스텀 대시보드는 Grafana UI에서 Import하거나 `grafana/dashboards/` 디렉토리에 JSON 파일로 추가할 수 있습니다.

## 보안 고려사항

1. **비밀번호 변경**
   - Grafana 관리자 비밀번호를 반드시 변경하세요
   - 프로덕션 환경에서는 Kubernetes Secret 사용을 권장합니다

2. **네트워크 접근 제어**
   - 방화벽 규칙을 적절히 설정하세요
   - Ingress/LoadBalancer 사용 시 TLS 설정을 권장합니다

3. **데이터 보안**
   - 민감한 정보는 환경 변수나 Secret으로 관리하세요
   - 로그에 민감한 정보가 포함되지 않도록 주의하세요

