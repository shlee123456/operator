# Prometheus + Grafana + Loki + Alloy on Kubernetes

Prometheus, Grafana, Loki, Alloy를 Kubernetes 클러스터에 배포하기 위한 매니페스트입니다.

## 구성 요소

- **Prometheus**: 메트릭 수집 및 저장 (포트: 9090)
- **Grafana**: 대시보드 및 시각화 (포트: 3000)
- **Loki**: 로그 수집 및 저장 (포트: 3100)
- **Alloy**: Grafana Agent (메트릭 및 로그 수집)

## 사전 요구사항

- Kubernetes 클러스터 (v1.20+)
- kubectl 설치 및 클러스터 접근 권한
- 스토리지 프로비저너 (동적 PV 프로비저닝용)

## 배포 방법

### 1. Kustomize를 사용한 배포 (권장)

```bash
# 프로젝트 디렉토리로 이동
cd /Users/shlee/Documents/Source/prometheus-grafana-k8s

# Kustomize로 배포
kubectl apply -k .
```

### 2. kubectl로 직접 배포

```bash
# Namespace 생성
kubectl apply -f namespace.yaml

# Prometheus 배포
kubectl apply -f prometheus/

# Grafana 배포
kubectl apply -f grafana/

# Loki 배포
kubectl apply -f loki/

# Alloy 배포
kubectl apply -f alloy/
```

## 배포 확인

```bash
# Pod 상태 확인
kubectl get pods -n monitoring

# Service 확인
kubectl get svc -n monitoring

# PVC 확인
kubectl get pvc -n monitoring
```

## 접근 방법

### Grafana 접근

Grafana는 LoadBalancer 타입으로 배포됩니다.

```bash
# External IP 확인
kubectl get svc grafana -n monitoring

# 포트 포워딩으로 로컬 접근 (LoadBalancer를 사용하지 않는 경우)
kubectl port-forward -n monitoring svc/grafana 3000:3000
```

- URL: http://localhost:3000 (포트 포워딩 사용 시)
- 기본 계정:
  - Username: `admin`
  - Password: `admin`

### Prometheus 접근

```bash
# 포트 포워딩
kubectl port-forward -n monitoring svc/prometheus 9090:9090
```

- URL: http://localhost:9090

### Loki 접근

```bash
# 포트 포워딩
kubectl port-forward -n monitoring svc/loki 3100:3100
```

- URL: http://localhost:3100

## 스토리지 설정

기본 스토리지 요청량:
- Prometheus: 10Gi
- Grafana: 5Gi
- Loki: 10Gi

스토리지 용량을 변경하려면 각 PVC 파일을 수정하세요:
- `prometheus/pvc.yaml`
- `grafana/pvc.yaml`
- `loki/pvc.yaml`

## 리소스 제한

각 컴포넌트의 리소스 설정:

| 컴포넌트 | CPU Request | CPU Limit | Memory Request | Memory Limit |
|---------|------------|-----------|----------------|-------------|
| Prometheus | 500m | 2000m | 512Mi | 2Gi |
| Grafana | 250m | 1000m | 256Mi | 1Gi |
| Loki | 500m | 2000m | 512Mi | 2Gi |
| Alloy | 250m | 1000m | 256Mi | 1Gi |

리소스를 조정하려면 각 deployment 파일을 수정하세요.

## 설정 변경

### Prometheus 설정 변경

`prometheus/configmap.yaml` 파일을 수정한 후:

```bash
kubectl apply -f prometheus/configmap.yaml
kubectl rollout restart deployment/prometheus -n monitoring
```

### Grafana 데이터소스 변경

`grafana/configmap-datasources.yaml` 파일을 수정한 후:

```bash
kubectl apply -f grafana/configmap-datasources.yaml
kubectl rollout restart deployment/grafana -n monitoring
```

### Loki 설정 변경

`loki/configmap.yaml` 파일을 수정한 후:

```bash
kubectl apply -f loki/configmap.yaml
kubectl rollout restart deployment/loki -n monitoring
```

### Alloy 설정 변경

`alloy/configmap.yaml` 파일을 수정한 후:

```bash
kubectl apply -f alloy/configmap.yaml
kubectl rollout restart deployment/alloy -n monitoring
```

## 삭제

```bash
# Kustomize 사용 시
kubectl delete -k .

# 또는 개별 삭제
kubectl delete namespace monitoring
```

## 문제 해결

### Pod가 Pending 상태인 경우

```bash
# Pod 상세 정보 확인
kubectl describe pod <pod-name> -n monitoring

# 이벤트 확인
kubectl get events -n monitoring --sort-by='.lastTimestamp'
```

보통 PVC 프로비저닝 문제이거나 리소스 부족일 수 있습니다.

### 로그 확인

```bash
# Prometheus 로그
kubectl logs -n monitoring deployment/prometheus

# Grafana 로그
kubectl logs -n monitoring deployment/grafana

# Loki 로그
kubectl logs -n monitoring deployment/loki

# Alloy 로그
kubectl logs -n monitoring deployment/alloy
```

## 주의사항

1. **보안**:
   - Grafana 관리자 비밀번호를 변경하세요 (`grafana/deployment.yaml`의 `GF_SECURITY_ADMIN_PASSWORD`)
   - 프로덕션 환경에서는 Secret 리소스를 사용하는 것을 권장합니다

2. **네트워킹**:
   - 현재 Grafana는 LoadBalancer 타입으로 설정되어 있습니다
   - 클라우드 환경이 아니라면 NodePort 또는 Ingress 사용을 고려하세요

3. **스토리지**:
   - PVC는 동적 프로비저닝을 사용합니다
   - 스토리지 클래스가 없다면 수동으로 PV를 생성해야 합니다

4. **외부 메트릭 수집**:
   - Prometheus 설정에 외부 Node Exporter 타겟이 포함되어 있습니다
   - 네트워크 접근이 가능한지 확인하세요 (예: 10.0.1.10:9100, 10.0.1.20:9100)

## 추가 설정

### Ingress 설정 (선택사항)

Grafana를 Ingress로 노출하려면:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: grafana-ingress
  namespace: monitoring
spec:
  rules:
  - host: grafana.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: grafana
            port:
              number: 3000
```

### 영구 스토리지 백업

```bash
# PVC 데이터 백업
kubectl cp monitoring/<prometheus-pod-name>:/prometheus ./prometheus-backup
kubectl cp monitoring/<grafana-pod-name>:/var/lib/grafana ./grafana-backup
kubectl cp monitoring/<loki-pod-name>:/loki ./loki-backup
```
