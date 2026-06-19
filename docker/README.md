# 통합 모니터링 스택 (Prometheus · Grafana · Loki · Alloy)

Docker Compose 기반의 즉시 적용 가능한 모니터링 스택입니다.
**다른 프로젝트에 복사해 넣은 뒤, 대상 컨테이너에 라벨만 붙이면 메트릭과 로그가 자동으로 수집**됩니다.

## 🎯 구성 요소

- **Prometheus**: 메트릭 수집 및 저장 (보존 30일)
- **Grafana**: 시각화 및 대시보드 (데이터소스·대시보드 자동 프로비저닝)
- **Loki**: 로그 수집 및 저장
- **Alloy**: 통합 에이전트 — 호스트 메트릭 수집 + Docker 컨테이너 로그 자동 수집 + 라벨 기반 앱 메트릭 수집

## ⚠️ 플랫폼 전제 (중요)

호스트 시스템 메트릭 수집(`prometheus.exporter.unix`, `/proc`·`/sys`·`/rootfs` 마운트)은 **Linux 호스트 기준**입니다.

- **Linux 서버**: 호스트 메트릭이 정확하게 수집됩니다. (권장 배포 환경)
- **macOS / Windows (Docker Desktop)**: `/proc`·`/sys`가 호스트가 아닌 내부 리눅스 VM을 가리키므로, 수집되는 호스트 메트릭은 Mac/Windows 본체가 아닌 VM 기준입니다. 컨테이너 로그·앱 메트릭 수집은 정상 동작하므로 로컬 개발·테스트 용도로는 문제없습니다.

## 🚀 빠른 시작

```bash
# 1) 환경 변수 설정 (선택, 기본값으로도 동작)
cp .env.example .env        # 필요 시 Grafana 비밀번호·이미지 버전 수정

# 2) 공용 네트워크 생성 (최초 1회)
docker network create shared_net

# 3) 스택 기동
docker compose up -d
```

기동 후 접속:

| 서비스 | URL | 비고 |
|--------|-----|------|
| Grafana | http://localhost:3000 | 기본 admin / admin (.env로 변경 가능) |
| Prometheus | http://localhost:9090 | |
| Loki | http://localhost:3100 | HTTP API |
| Alloy | http://localhost:12345 | 에이전트 상태 UI |

## 📈 다른 프로젝트에 적용하는 법

### 로그 — 추가 설정 불필요
Alloy가 Docker 소켓을 통해 **실행 중인 모든 컨테이너의 stdout/stderr 로그를 자동 수집**하여 Loki로 보냅니다.
대상 프로젝트의 컨테이너가 떠 있기만 하면 Grafana의 **"로그 분석 > 컨테이너 로그"** 대시보드에서 바로 확인됩니다.

### 애플리케이션 메트릭 — 라벨만 추가
대상 프로젝트의 `docker-compose.yml`에서 메트릭을 노출하는 컨테이너에 라벨을 붙이고, `shared_net`에 연결합니다.

```yaml
services:
  my-app:
    image: my-app:latest
    labels:
      - "prometheus.io/scrape=true"   # (필수) 수집 활성화
      - "prometheus.io/port=8080"      # (필수) 메트릭 포트
      - "prometheus.io/path=/metrics"  # (선택) 기본값 /metrics
    networks:
      - shared_net

networks:
  shared_net:
    external: true
```

이렇게만 하면 Alloy가 자동으로 발견해 수집하며, Grafana **"애플리케이션 모니터링 > 수집 대상 상태"** 대시보드에서 UP/DOWN과 수집 상태를 확인할 수 있습니다. `prometheus.yml`은 수정할 필요가 없습니다.

> 라벨을 쓸 수 없는 원격 서버/외부 엔드포인트는 `prometheus/prometheus.yml`의 `scrape_configs`에 정적 타겟으로 직접 추가하세요.

## 📊 기본 제공 대시보드

| 폴더 | 대시보드 | 데이터소스 |
|------|----------|-----------|
| 인프라 모니터링 | 시스템 개요 (CPU/메모리/디스크) | Prometheus |
| 애플리케이션 모니터링 | 수집 대상 상태 (UP/DOWN, 수집 시간) | Prometheus |
| 로그 분석 | 컨테이너 로그 (발생량/에러/스트림, 컨테이너 필터) | Loki |

## 🔧 버전 관리

이미지 버전은 `latest`가 아닌 고정 버전을 사용합니다(재현성). 변경은 `.env`에서:

```dotenv
PROMETHEUS_VERSION=v3.1.0
GRAFANA_VERSION=11.4.0
LOKI_VERSION=3.3.2
ALLOY_VERSION=v1.5.1
```

## 🛠️ 운영 명령어

```bash
docker compose up -d                 # 기동
docker compose down                  # 중지
docker compose logs -f alloy         # 특정 서비스 로그
docker compose restart prometheus    # 설정 변경 후 재시작
docker network create shared_net     # 네트워크 최초 생성
```

## 💾 데이터 영속성

- Grafana: `grafana-storage` 볼륨에 영구 저장
- Prometheus / Loki: 기본은 컨테이너 내부 (영속화하려면 `docker-compose.yml`의 볼륨 주석 해제)

## 🧩 문제 해결

- **`network shared_net not found`**: `docker network create shared_net` 실행
- **앱 메트릭이 안 잡힘**: 대상 컨테이너가 `shared_net`에 연결됐는지, `prometheus.io/scrape=true`와 `prometheus.io/port`가 정확한지 확인. Alloy UI(:12345)의 `discovery.relabel "app_metrics"` 출력 타겟 확인
- **포트 충돌**: `lsof -i :3000` 등으로 확인 후 `docker-compose.yml`의 포트 매핑 변경
