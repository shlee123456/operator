# Kubernetes 배포 CLAUDE.md

> **상위 문서**: [루트 CLAUDE.md](../CLAUDE.md)를 먼저 참조하세요.
> 이 문서는 루트 규칙을 따르며, Kubernetes 배포에 특화된 규칙만 정의합니다.

## 목적

Kubernetes 클러스터에 모니터링 스택을 배포하기 위한 매니페스트 및 배포 규칙을 정의합니다.

## 디렉토리 구조

```
k8s/
├── kustomization.yaml      # Kustomize 설정
├── namespace.yaml          # monitoring 네임스페이스
├── prometheus/
│   ├── configmap.yaml      # Prometheus 설정
│   ├── deployment.yaml     # Prometheus 배포
│   ├── service.yaml        # Prometheus 서비스
│   └── pvc.yaml            # Prometheus 스토리지
├── grafana/
│   ├── configmap-datasources.yaml  # Grafana 데이터소스
│   ├── configmap-dashboards.yaml  # Grafana 대시보드
│   ├── deployment.yaml     # Grafana 배포
│   ├── service.yaml        # Grafana 서비스
│   ├── ingress.yaml        # Grafana Ingress (선택)
│   └── pvc.yaml            # Grafana 스토리지
├── loki/
│   ├── configmap.yaml      # Loki 설정
│   ├── deployment.yaml     # Loki 배포
│   ├── service.yaml        # Loki 서비스
│   └── pvc.yaml            # Loki 스토리지
├── alloy/
│   ├── configmap.yaml      # Alloy 설정
│   ├── deployment.yaml     # Alloy 배포
│   ├── service.yaml        # Alloy 서비스
│   └── rbac.yaml           # Alloy RBAC 권한
├── examples/               # 예제 및 참고 자료
│   ├── alerts/            # 알림 설정 예제
│   ├── exporters/         # Exporter 배포 예제
│   └── app-examples/      # 애플리케이션 메트릭 예제
├── deploy.sh              # 배포 스크립트
├── undeploy.sh            # 삭제 스크립트
└── README.md
```

## 로컬 코딩 컨벤션

### Kubernetes 매니페스트
- 모든 리소스는 `monitoring` 네임스페이스 사용
- 리소스 이름: kebab-case (예: `prometheus`, `grafana-service`)
- ConfigMap: 설정 파일은 ConfigMap으로 관리
- PVC: 동적 프로비저닝 사용 (스토리지 클래스 필요)
- 리소스 제한: 모든 Deployment에 requests/limits 설정

### Kustomize 사용
- `kustomization.yaml`로 모든 리소스 통합 관리
- 배포: `kubectl apply -k .`
- 삭제: `kubectl delete -k .`

## 주요 파일

| 파일 | 설명 |
|------|------|
| kustomization.yaml | Kustomize 설정 (모든 리소스 통합) |
| namespace.yaml | monitoring 네임스페이스 정의 |
| prometheus/configmap.yaml | Prometheus 설정 (스크랩 타겟 등) |
| prometheus/deployment.yaml | Prometheus Pod 배포 |
| prometheus/pvc.yaml | Prometheus 데이터 스토리지 (10Gi) |
| grafana/configmap-datasources.yaml | Grafana 데이터소스 자동 설정 |
| grafana/configmap-dashboards.yaml | Grafana 대시보드 자동 프로비저닝 |
| grafana/deployment.yaml | Grafana Pod 배포 |
| grafana/pvc.yaml | Grafana 데이터 스토리지 (5Gi) |
| loki/configmap.yaml | Loki 설정 |
| loki/deployment.yaml | Loki Pod 배포 |
| loki/pvc.yaml | Loki 데이터 스토리지 (10Gi) |
| alloy/configmap.yaml | Alloy 설정 |
| alloy/deployment.yaml | Alloy Pod 배포 |
| alloy/rbac.yaml | Alloy Kubernetes 권한 (메트릭 수집용) |

## 자주 사용하는 명령어

```bash
# 배포
kubectl apply -k .

# 삭제
kubectl delete -k .

# 또는 스크립트 사용
./deploy.sh
./undeploy.sh

# 상태 확인
kubectl get pods -n monitoring
kubectl get svc -n monitoring
kubectl get pvc -n monitoring

# 로그 확인
kubectl logs -n monitoring deployment/prometheus
kubectl logs -n monitoring deployment/grafana
kubectl logs -n monitoring deployment/loki
kubectl logs -n monitoring deployment/alloy

# 포트 포워딩
kubectl port-forward -n monitoring svc/grafana 3000:3000
kubectl port-forward -n monitoring svc/prometheus 9090:9090
kubectl port-forward -n monitoring svc/loki 3100:3100

# 설정 변경 후 재시작
kubectl rollout restart deployment/prometheus -n monitoring
kubectl rollout restart deployment/grafana -n monitoring
kubectl rollout restart deployment/loki -n monitoring
kubectl rollout restart deployment/alloy -n monitoring
```

## 설정 변경 시 주의사항

### Prometheus 설정 변경
1. `prometheus/configmap.yaml` 수정
2. ConfigMap 적용: `kubectl apply -f prometheus/configmap.yaml`
3. Deployment 재시작: `kubectl rollout restart deployment/prometheus -n monitoring`

### Grafana 설정 변경
1. `grafana/configmap-datasources.yaml` 또는 `grafana/configmap-dashboards.yaml` 수정
2. ConfigMap 적용: `kubectl apply -f grafana/configmap-*.yaml`
3. Deployment 재시작: `kubectl rollout restart deployment/grafana -n monitoring`

### Loki 설정 변경
1. `loki/configmap.yaml` 수정
2. ConfigMap 적용: `kubectl apply -f loki/configmap.yaml`
3. Deployment 재시작: `kubectl rollout restart deployment/loki -n monitoring`

### Alloy 설정 변경
1. `alloy/configmap.yaml` 수정
2. ConfigMap 적용: `kubectl apply -f alloy/configmap.yaml`
3. Deployment 재시작: `kubectl rollout restart deployment/alloy -n monitoring`

## 리소스 제한

기본 리소스 설정:

| 컴포넌트 | CPU Request | CPU Limit | Memory Request | Memory Limit |
|---------|------------|-----------|----------------|-------------|
| Prometheus | 500m | 2000m | 512Mi | 2Gi |
| Grafana | 250m | 1000m | 256Mi | 1Gi |
| Loki | 500m | 2000m | 512Mi | 2Gi |
| Alloy | 250m | 1000m | 256Mi | 1Gi |

리소스 조정은 각 `deployment.yaml` 파일에서 수정합니다.

## 스토리지 설정

기본 스토리지 요청량:
- Prometheus: 10Gi
- Grafana: 5Gi
- Loki: 10Gi

스토리지 용량 변경은 각 `pvc.yaml` 파일에서 수정합니다.

**주의**: 동적 프로비저닝을 사용하므로 스토리지 클래스가 필요합니다. 스토리지 클래스가 없는 경우:
1. PVC를 수동으로 생성하거나
2. `pvc.yaml`에서 `storageClassName`을 기존 스토리지 클래스로 변경

## 네트워크 설정

### Service 타입
- Prometheus: ClusterIP (내부 접근)
- Grafana: LoadBalancer (외부 접근, 클라우드 환경)
- Loki: ClusterIP (내부 접근)
- Alloy: ClusterIP (내부 접근)

### Ingress 설정 (선택)
- `grafana/ingress.yaml` 파일이 포함되어 있습니다
- Ingress Controller가 설치되어 있어야 합니다
- 도메인 및 TLS 설정을 수정하여 사용하세요

## RBAC 설정

Alloy는 Kubernetes 메트릭을 수집하기 위해 RBAC 권한이 필요합니다:
- `alloy/rbac.yaml`에 ServiceAccount, ClusterRole, ClusterRoleBinding 정의
- Pod 메트릭, Node 메트릭, Service 메트릭 수집 권한 포함

## 문제 해결

### Pod가 Pending 상태
```bash
# Pod 상세 정보 확인
kubectl describe pod <pod-name> -n monitoring

# 이벤트 확인
kubectl get events -n monitoring --sort-by='.lastTimestamp'

# 일반적인 원인:
# - PVC 프로비저닝 실패 (스토리지 클래스 없음)
# - 리소스 부족 (노드에 충분한 CPU/메모리 없음)
```

### PVC가 Pending 상태
```bash
# PVC 상세 정보 확인
kubectl describe pvc <pvc-name> -n monitoring

# 스토리지 클래스 확인
kubectl get storageclass

# 동적 프로비저닝이 안 되는 경우 수동 PV 생성 필요
```

### 설정 변경이 반영되지 않음
```bash
# ConfigMap이 제대로 적용되었는지 확인
kubectl get configmap -n monitoring
kubectl describe configmap <configmap-name> -n monitoring

# Deployment 재시작
kubectl rollout restart deployment/<deployment-name> -n monitoring

# Pod 로그 확인
kubectl logs -n monitoring deployment/<deployment-name>
```

## 예제 및 참고 자료

`examples/` 디렉토리에는 다음 예제가 포함되어 있습니다:
- `alerts/`: Alertmanager 및 Prometheus 알림 규칙 예제
- `exporters/`: 다양한 Exporter 배포 예제 (Node Exporter, Postgres Exporter 등)
- `app-examples/`: 애플리케이션에서 메트릭을 노출하는 예제 코드

## 보안 고려사항

1. **비밀번호 관리**
   - Grafana 관리자 비밀번호는 Secret으로 관리 권장
   - 현재는 환경 변수로 설정되어 있음 (`grafana/deployment.yaml`)

2. **네트워크 정책**
   - 프로덕션 환경에서는 NetworkPolicy 적용 권장
   - 불필요한 포트 노출 방지

3. **RBAC 최소 권한 원칙**
   - Alloy RBAC는 필요한 권한만 부여
   - 추가 권한이 필요한 경우 신중하게 검토
