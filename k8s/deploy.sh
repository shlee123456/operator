#!/bin/bash

set -e

echo "=========================================="
echo "Prometheus + Grafana + Loki + Alloy 배포"
echo "=========================================="
echo ""

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 함수 정의
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        echo -e "${RED}Error: kubectl이 설치되어 있지 않습니다.${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ kubectl 확인 완료${NC}"
}

check_cluster() {
    if ! kubectl cluster-info &> /dev/null; then
        echo -e "${RED}Error: Kubernetes 클러스터에 연결할 수 없습니다.${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Kubernetes 클러스터 연결 확인${NC}"
}

deploy_resources() {
    echo ""
    echo "리소스 배포 중..."

    # Kustomize를 사용한 배포
    if command -v kustomize &> /dev/null || kubectl kustomize --help &> /dev/null; then
        echo -e "${YELLOW}Kustomize를 사용하여 배포합니다...${NC}"
        kubectl apply -k .
    else
        echo -e "${YELLOW}Kustomize를 사용할 수 없습니다. kubectl로 직접 배포합니다...${NC}"

        # Namespace
        kubectl apply -f namespace.yaml

        # Prometheus
        kubectl apply -f prometheus/

        # Grafana
        kubectl apply -f grafana/

        # Loki
        kubectl apply -f loki/

        # Alloy
        kubectl apply -f alloy/
    fi

    echo -e "${GREEN}✓ 리소스 배포 완료${NC}"
}

wait_for_pods() {
    echo ""
    echo "Pod가 실행될 때까지 대기 중..."

    # 최대 5분 대기
    kubectl wait --for=condition=ready pod -l app=prometheus -n monitoring --timeout=300s || true
    kubectl wait --for=condition=ready pod -l app=grafana -n monitoring --timeout=300s || true
    kubectl wait --for=condition=ready pod -l app=loki -n monitoring --timeout=300s || true
    kubectl wait --for=condition=ready pod -l app=alloy -n monitoring --timeout=300s || true
}

show_status() {
    echo ""
    echo "=========================================="
    echo "배포 상태"
    echo "=========================================="
    echo ""

    echo "Pods:"
    kubectl get pods -n monitoring
    echo ""

    echo "Services:"
    kubectl get svc -n monitoring
    echo ""

    echo "PersistentVolumeClaims:"
    kubectl get pvc -n monitoring
    echo ""
}

show_access_info() {
    echo "=========================================="
    echo "접근 정보"
    echo "=========================================="
    echo ""

    echo -e "${GREEN}Grafana:${NC}"
    echo "  URL: http://localhost:3000 (포트 포워딩 필요)"
    echo "  Username: admin"
    echo "  Password: admin"
    echo ""
    echo "  포트 포워딩 명령어:"
    echo "  kubectl port-forward -n monitoring svc/grafana 3000:3000"
    echo ""

    echo -e "${GREEN}Prometheus:${NC}"
    echo "  URL: http://localhost:9090 (포트 포워딩 필요)"
    echo ""
    echo "  포트 포워딩 명령어:"
    echo "  kubectl port-forward -n monitoring svc/prometheus 9090:9090"
    echo ""

    echo -e "${GREEN}Loki:${NC}"
    echo "  URL: http://localhost:3100 (포트 포워딩 필요)"
    echo ""
    echo "  포트 포워딩 명령어:"
    echo "  kubectl port-forward -n monitoring svc/loki 3100:3100"
    echo ""
}

# 메인 실행
main() {
    check_kubectl
    check_cluster
    deploy_resources
    wait_for_pods
    show_status
    show_access_info

    echo -e "${GREEN}=========================================="
    echo "배포가 완료되었습니다!"
    echo "==========================================${NC}"
}

# 스크립트 실행
main
