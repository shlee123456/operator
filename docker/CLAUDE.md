# Docker 배포 CLAUDE.md

> **상위 문서**: [루트 CLAUDE.md](../CLAUDE.md)를 먼저 참조하세요.
> 이 문서는 루트 규칙을 따르며, Docker Compose 배포에 특화된 규칙만 정의합니다.

## 목적

Docker Compose를 사용한 로컬 개발 및 테스트 환경 구축을 위한 설정 및 규칙을 정의합니다.

## 디렉토리 구조

```
docker/
├── docker-compose.yml      # Docker Compose 설정 (이미지 버전 고정, env 변수화)
├── .env.example            # 환경 변수 예시 (.env로 복사하여 사용)
├── prometheus/
│   └── prometheus.yml      # Prometheus 설정 (정적 타겟용, 보조)
├── grafana/
│   └── provisioning/      # Grafana 자동 프로비저닝 설정
│       ├── datasources/   # 데이터소스 (uid: prometheus, loki 고정)
│       └── dashboards/
│           ├── infrastructure/  # 시스템 개요
│           ├── application/     # 수집 대상 상태
│           └── logs/            # 컨테이너 로그
├── loki/
│   └── loki.yml           # Loki 설정
├── alloy/
│   └── config.alloy       # Alloy 설정 (호스트 메트릭 + 컨테이너 로그 + 라벨 기반 앱 메트릭)
└── README.md
```

## 로컬 코딩 컨벤션

### Docker Compose 파일
- 서비스별로 섹션 구분 (주석으로 명확히 표시)
- 환경 변수는 `.env` 파일 사용 (gitignore에 포함)
- 볼륨 마운트는 상대 경로 사용
- 네트워크는 `shared_net` 사용 (외부 네트워크)

### 설정 파일
- Prometheus: `prometheus/prometheus.yml`
- Loki: `loki/loki.yml`
- Alloy: `alloy/config.alloy`
- 모든 설정 파일은 주석을 통한 상세 설명 포함

## 주요 파일

| 파일 | 설명 |
|------|------|
| docker-compose.yml | 전체 스택 정의. 이미지 버전은 `.env`의 `*_VERSION`으로 고정 |
| .env.example | Grafana 계정·이미지 버전 변수. `.env`로 복사해 사용 (.env는 gitignore) |
| prometheus/prometheus.yml | 정적 타겟용 보조 설정 (앱 메트릭은 Alloy 라벨 방식 권장) |
| loki/loki.yml | 로그 수집 설정 |
| alloy/config.alloy | 통합 에이전트. 호스트 메트릭 + 컨테이너 로그 자동수집 + 라벨 기반 앱 메트릭 |
| grafana/provisioning/ | 데이터소스(uid 고정) + 대시보드 자동 프로비저닝 |

## 애플리케이션 메트릭 수집 (라벨 기반)

대상 컨테이너에 아래 라벨을 붙이고 `shared_net`에 연결하면 Alloy가 자동 수집한다.
`prometheus.yml` 수정 불필요. 상세는 `alloy/config.alloy`의 `discovery.relabel "app_metrics"` 참조.

```yaml
labels:
  - "prometheus.io/scrape=true"   # 필수
  - "prometheus.io/port=8080"      # 필수
  - "prometheus.io/path=/metrics"  # 선택
```

## 플랫폼 주의

호스트 메트릭(`prometheus.exporter.unix`)은 Linux 호스트 기준. macOS/Windows의 Docker Desktop에서는 호스트 본체가 아닌 리눅스 VM 메트릭이 수집된다. 로그·앱 메트릭은 정상 동작.

## 자주 사용하는 명령어

```bash
# 서비스 시작
docker-compose up -d

# 서비스 중지
docker-compose down

# 특정 서비스 로그 확인
docker-compose logs -f prometheus
docker-compose logs -f grafana
docker-compose logs -f loki
docker-compose logs -f alloy

# 서비스 재시작
docker-compose restart [service-name]

# 네트워크 생성 (처음 실행 시)
docker network create shared_net

# 네트워크 확인
docker network ls | grep shared_net
```

## 설정 변경 시 주의사항

### Prometheus 설정 변경
1. `prometheus/prometheus.yml` 수정
2. Prometheus 컨테이너 재시작: `docker-compose restart prometheus`
3. 또는 전체 재시작: `docker-compose restart`

### Grafana 설정 변경
1. `grafana/provisioning/` 디렉토리 내 파일 수정
2. Grafana 컨테이너 재시작: `docker-compose restart grafana`

### Loki 설정 변경
1. `loki/loki.yml` 수정
2. Loki 컨테이너 재시작: `docker-compose restart loki`

### Alloy 설정 변경
1. `alloy/config.alloy` 수정
2. Alloy 컨테이너 재시작: `docker-compose restart alloy`

## 데이터 영속성

- Grafana 데이터: `grafana-storage` 볼륨 사용
- Prometheus 데이터: 기본적으로 컨테이너 내부 (볼륨 주석 처리됨)
- Loki 데이터: 기본적으로 컨테이너 내부 (볼륨 주석 처리됨)

데이터 영속성이 필요한 경우 `docker-compose.yml`의 볼륨 주석을 해제하세요.

## 네트워크 설정

- 모든 서비스는 `shared_net` 네트워크 사용
- 외부 네트워크로 설정되어 있으므로 사전에 생성 필요:
  ```bash
  docker network create shared_net
  ```
- 내부 네트워크를 사용하려면 `docker-compose.yml`에서 `external: true` 제거

## 포트 매핑

| 서비스 | 호스트 포트 | 컨테이너 포트 | 용도 |
|--------|------------|--------------|------|
| Prometheus | 9090 | 9090 | Web UI 및 API |
| Grafana | 3000 | 3000 | Web UI |
| Loki | 3100 | 3100 | HTTP API |
| Alloy | 12345 | 12345 | Web UI 및 메트릭 엔드포인트 |

## 문제 해결

### 네트워크 오류
```bash
# 네트워크가 없는 경우
docker network create shared_net

# 네트워크 확인
docker network inspect shared_net
```

### 포트 충돌
```bash
# 포트 사용 중인 프로세스 확인
lsof -i :9090
lsof -i :3000
lsof -i :3100
lsof -i :12345

# docker-compose.yml에서 포트 변경
```

### 볼륨 권한 문제
```bash
# 볼륨 권한 확인
docker-compose down
sudo chown -R $USER:$USER grafana-storage/
docker-compose up -d
```
