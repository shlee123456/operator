#!/bin/bash

set -e

echo "=========================================="
echo "Prometheus + Grafana + Loki + Alloy 삭제"
echo "=========================================="
echo ""

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 확인 메시지
echo -e "${YELLOW}경고: 모든 모니터링 리소스가 삭제됩니다.${NC}"
echo -e "${YELLOW}데이터가 영구적으로 삭제될 수 있습니다.${NC}"
echo ""
read -p "계속하시겠습니까? (y/N): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "삭제가 취소되었습니다."
    exit 0
fi

echo ""
echo "리소스 삭제 중..."

# Kustomize를 사용한 삭제
if command -v kustomize &> /dev/null || kubectl kustomize --help &> /dev/null; then
    echo -e "${YELLOW}Kustomize를 사용하여 삭제합니다...${NC}"
    kubectl delete -k . --ignore-not-found=true
else
    echo -e "${YELLOW}kubectl로 직접 삭제합니다...${NC}"

    # 순서대로 삭제
    kubectl delete -f alloy/ --ignore-not-found=true
    kubectl delete -f loki/ --ignore-not-found=true
    kubectl delete -f grafana/ --ignore-not-found=true
    kubectl delete -f prometheus/ --ignore-not-found=true
    kubectl delete -f namespace.yaml --ignore-not-found=true
fi

echo -e "${GREEN}✓ 리소스 삭제 완료${NC}"
echo ""
echo -e "${GREEN}=========================================="
echo "삭제가 완료되었습니다!"
echo "==========================================${NC}"
