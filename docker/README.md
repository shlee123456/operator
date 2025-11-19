# Prometheus & Grafana 모니터링 시스템

Docker Compose를 사용하여 Prometheus와 Grafana 기반의 서버 모니터링 시스템을 구축하는 프로젝트입니다.

## 🎯 프로젝트 개요

이 프로젝트는 여러 서버의 시스템 메트릭을 실시간으로 수집하고 시각화하기 위한 모니터링 솔루션입니다.

### 주요 구성 요소

- **Prometheus**: 메트릭 수집 및 저장
- **Grafana**: 데이터 시각화 및 대시보드
- **Node Exporter**: 각 서버의 시스템 메트릭 수집 (별도 설치 필요)

## 📁 프로젝트 구조

```
prometheus-grafana/
├── docker-compose.yml      # Docker Compose 설정 파일
├── prometheus/
│   └── prometheus.yml      # Prometheus 설정 파일
└── README.md              # 프로젝트 문서
```

## 🚀 빠른 시작

### 사전 요구사항

- Docker
- Docker Compose
- 모니터링 대상 서버에 Node Exporter 설치

### 설치 및 실행

1. **저장소 클론**
   ```bash
   git clone <repository-url>
   cd prometheus-grafana
   ```

2. **Docker 네트워크 생성** (처음 실행 시)
   ```bash
   docker network create shared_net
   ```

3. **서비스 시작**
   ```bash
   docker-compose up -d
   ```

## 🌐 접속 정보

### Prometheus
- **URL**: http://localhost:9090
- **설명**: 메트릭 수집 상태 확인 및 쿼리 테스트

### Grafana
- **URL**: http://localhost:3000
- **기본 계정**: admin / admin
- **설명**: 대시보드를 통한 데이터 시각화

## ⚙️ 설정

### Prometheus 설정

`prometheus/prometheus.yml` 파일에서 모니터링 대상 서버를 설정합니다.

```yaml
scrape_configs:
  - job_name: 'node_exporter'
    static_configs:
      - targets: ['<서버IP>:9100']
        labels:
          instance: '<서버IP>'
          domainname: '<서버명>'
```

### 현재 모니터링 대상

- **Demo 서버들**: demoA, demoB, demoC, demoD
- **의료기관**: 강릉아산병원, 화순전남대병원
- **DB 서버**: DB_Active, DB_Stanby, 백업서버
- **기타**: DEAN, voice-emrs, core.puzzle-ai.com

## 📊 주요 기능

- **실시간 시스템 모니터링**: CPU, 메모리, 디스크, 네트워크 사용량
- **알림 시스템**: 임계값 초과 시 알림 (별도 설정 필요)
- **데이터 보존**: 30일간 메트릭 데이터 저장
- **멀티 서버 관리**: 여러 서버를 통합 모니터링

## 🔧 유지보수

### 서비스 관리

```bash
# 서비스 중지
docker-compose down

# 서비스 재시작
docker-compose restart

# 로그 확인
docker-compose logs -f
```

### 데이터 관리

- Prometheus 데이터는 30일 후 자동 삭제됩니다
- Grafana 데이터는 `grafana-storage` 볼륨에 영구 저장됩니다

## 📝 추가 설정

### Node Exporter 설치 (모니터링 대상 서버)

각 모니터링 대상 서버에 Node Exporter를 설치해야 합니다:

```bash
# 다운로드 및 설치
wget https://github.com/prometheus/node_exporter/releases/download/v1.6.1/node_exporter-1.6.1.linux-amd64.tar.gz
tar xvfz node_exporter-1.6.1.linux-amd64.tar.gz
sudo mv node_exporter-1.6.1.linux-amd64/node_exporter /usr/local/bin/

# 서비스로 등록
sudo systemctl enable node_exporter
sudo systemctl start node_exporter
```

### Grafana 대시보드

1. Grafana에 로그인 후 Data Source 추가
2. Prometheus를 Data Source로 설정 (http://prometheus:9090)
3. 대시보드 템플릿 import 또는 커스텀 대시보드 생성

### 로그 확인

```bash
# 특정 서비스 로그
docker-compose logs prometheus
docker-compose logs grafana
```

