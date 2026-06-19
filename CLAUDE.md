# Prometheus Monitoring Stack

Prometheus, Grafana, Loki, Alloy를 활용한 통합 모니터링 스택입니다. Docker와 Kubernetes 두 가지 배포 방식을 지원합니다.

## ⚠️ 필수 실행 규칙 (모든 작업 전 확인)

> **이 규칙은 모든 작업에서 반드시 준수해야 합니다.**

### 1. 터미널 로그 기록
모든 Bash 명령어(빌드, 테스트, 설치, 실행) 실행 시 `.context/terminal/`에 로그 저장:
```bash
[명령어] 2>&1 | tee .context/terminal/[명령어]_$(date +%s).log
```

### 2. 서브 CLAUDE.md 관리
- 새 디렉토리/모듈 생성 시 → 해당 디렉토리에 서브 CLAUDE.md 생성
- 기존 구조 변경 시 → 관련 서브 CLAUDE.md 업데이트
- 서브 CLAUDE.md는 반드시 루트 CLAUDE.md 참조

### 3. 브랜치 컨텍스트 참조
- **브랜치별 작업 지침**: `.context/branch/[브랜치명].md` 파일 확인
- 세션 시작 시 현재 브랜치에 해당하는 컨텍스트 파일이 있으면 반드시 참조
- 브랜치 컨텍스트에는 작업 목표, 범위 제한, 금지 사항 등이 정의됨

```bash
# 브랜치 컨텍스트 파일 확인
git branch --show-current  # 현재 브랜치 확인
cat .context/branch/$(git branch --show-current | tr '/' '-').md  # 컨텍스트 확인
```

### 4. 세션 관리
- **세션 시작**: `.context/history/`에서 최근 기록 확인, 이전 TODO 파악
- **세션 종료**: `.context/history/session_YYYY-MM-DD_HH-MM.md`에 작업 내용 기록

### 5. Git 커밋 규칙
- 커밋 메시지는 **한글**로 작성
- `Co-Authored-By` 태그 **사용 금지** (Claude 공동 작성자 표기 안 함)
- 커밋 메시지 형식:
```
<type>: <한글 설명>

<본문 (선택)>
```
- type: feat, fix, docs, refactor, test, chore 등

### 6. 작업 완료 체크리스트
- [ ] 터미널 로그 저장했는가?
- [ ] 서브 CLAUDE.md 업데이트 필요한가?
- [ ] 브랜치 컨텍스트 범위 내에서 작업했는가?
- [ ] 세션 히스토리 기록했는가?
- [ ] Git 커밋 시 한글 메시지 사용했는가?

---

## 프로젝트 개요

Prometheus, Grafana, Loki, Alloy로 구성된 통합 모니터링 스택입니다. Docker Compose와 Kubernetes 두 가지 배포 방식을 제공합니다.

## 기술 스택

- **메트릭 수집**: Prometheus
- **시각화**: Grafana
- **로그 수집**: Loki
- **통합 에이전트**: Alloy (Grafana Agent)
- **배포 방식**: Docker Compose, Kubernetes (Kustomize)

## 전역 코딩 컨벤션

- YAML 파일: 2 spaces 들여쓰기
- 설정 파일: 주석을 통한 상세 설명 포함
- 네이밍: kebab-case (파일명), camelCase (변수명)
- Kubernetes 리소스: 네임스페이스 `monitoring` 사용

## 서브 CLAUDE.md 목록

| 경로 | 설명 |
|------|------|
| docker/CLAUDE.md | Docker Compose 배포 관련 규칙 |
| k8s/CLAUDE.md | Kubernetes 배포 관련 규칙 |

## 세션 관리 규칙

### 세션 시작 시
1. `.context/history/`에서 최근 세션 파일 확인
2. 이전 세션의 진행상황과 TODO 파악
3. 중단된 작업이 있으면 이어서 진행

### 세션 종료 시
1. `.context/history/session_YYYY-MM-DD_HH-MM.md` 파일 생성
2. 다음 내용 기록:
   - 완료한 작업
   - 주요 결정사항
   - 다음 세션 TODO
   - 발생한 이슈와 해결 방법

## 자주 사용하는 명령어

### Docker Compose
```bash
# 서비스 시작
cd docker && docker-compose up -d

# 서비스 중지
cd docker && docker-compose down

# 로그 확인
cd docker && docker-compose logs -f [service-name]

# 네트워크 생성 (처음 실행 시)
docker network create shared_net
```

### Kubernetes
```bash
# 배포
cd k8s && kubectl apply -k .

# 삭제
cd k8s && kubectl delete -k .

# 상태 확인
kubectl get pods -n monitoring
kubectl get svc -n monitoring

# 포트 포워딩
kubectl port-forward -n monitoring svc/grafana 3000:3000
kubectl port-forward -n monitoring svc/prometheus 9090:9090
kubectl port-forward -n monitoring svc/loki 3100:3100
```

## 명령어 로그 기록 (필수)

```bash
# Docker Compose
docker-compose up -d 2>&1 | tee .context/terminal/docker-compose-up_$(date +%s).log
docker-compose logs -f 2>&1 | tee .context/terminal/docker-compose-logs_$(date +%s).log

# Kubernetes
kubectl apply -k . 2>&1 | tee .context/terminal/kubectl-apply_$(date +%s).log
kubectl get pods -n monitoring 2>&1 | tee .context/terminal/kubectl-status_$(date +%s).log
```

## 히스토리 관리

- `.context/history/`에 최근 5개 세션만 유지
- 7일 이상 된 히스토리는 `.context/archive/`로 이동
- 30일 이상 된 아카이브는 삭제

## 터미널 로그 관리

- `.context/terminal/`에 최근 10개 로그만 유지
- 오래된 로그는 자동 삭제

## 접속 정보

### Docker 환경
- Grafana: http://localhost:3000 (admin/admin)
- Prometheus: http://localhost:9090
- Loki: http://localhost:3100
- Alloy: http://localhost:12345

### Kubernetes 환경
- Grafana: 포트 포워딩 후 http://localhost:3000 (admin/admin)
- Prometheus: 포트 포워딩 후 http://localhost:9090
- Loki: 포트 포워딩 후 http://localhost:3100
