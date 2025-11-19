# Docker Desktop Kubernetes에 원격 노드 추가 (비권장)

## 주의사항

Docker Desktop Kubernetes는 단일 노드 개발 환경으로 설계되었습니다. 원격 워커 노드를 추가하는 것은:
- ❌ 공식적으로 지원되지 않음
- ❌ 프로덕션 환경에 부적합
- ❌ 네트워킹 문제 발생 가능
- ❌ 맥북 재시작 시 불안정

**권장: 대안 솔루션 사용 (아래 참조)**

---

## 이론적인 방법 (실험용)

### 1단계: Docker Desktop Kubernetes 정보 확인

```bash
# Kubernetes API 서버 주소 확인
kubectl cluster-info

# Control Plane 노드 확인
kubectl get nodes

# 토큰 생성 (워커 노드 조인용)
kubeadm token create --print-join-command
```

**문제점:** Docker Desktop K8s는 kubeadm으로 설치되지 않아서 위 명령어가 작동하지 않을 수 있습니다.

### 2단계: API 서버 외부 접근 설정

Docker Desktop의 API 서버는 기본적으로 localhost에만 바인딩되어 있습니다.

```bash
# 현재 컨텍스트 확인
kubectl config view

# API 서버 주소 확인 (보통 https://127.0.0.1:6443)
kubectl config view -o jsonpath='{.clusters[0].cluster.server}'
```

**문제점:** 맥북의 내부 IP로 접근하려면 포트 포워딩 및 방화벽 설정이 필요합니다.

### 3단계: 네트워킹 설정 (복잡함)

```bash
# 맥북의 IP 주소 확인
ifconfig | grep "inet " | grep -v 127.0.0.1

# 방화벽에서 Kubernetes 포트 열기
# - 6443: API Server
# - 10250: Kubelet
# - 10251: Kube-scheduler
# - 10252: Kube-controller-manager
# - 2379-2380: etcd
```

### 4단계: 원격 서버 준비

원격 서버에서:

```bash
# 1. Docker 설치
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 2. Kubernetes 패키지 설치
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl

curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.28/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.28/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl

# 3. 스왑 비활성화
sudo swapoff -a
sudo sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab
```

### 5단계: 조인 시도 (실패할 가능성 높음)

```bash
# 맥북에서 토큰 및 CA 해시 가져오기
# (Docker Desktop K8s에서는 이 정보를 얻기 어려움)

# 원격 서버에서
sudo kubeadm join <맥북-IP>:6443 --token <token> \
    --discovery-token-ca-cert-hash sha256:<hash>
```

**예상 문제:**
- API 서버 접근 불가
- 인증서 검증 실패
- 네트워크 정책 충돌

---

## 권장 대안 솔루션

### 옵션 1: K3s (가장 추천)

경량 Kubernetes로 멀티 노드 클러스터를 쉽게 구성할 수 있습니다.

**맥북 (Control Plane):**
```bash
# K3s 서버 설치 (맥북에서는 Rancher Desktop 또는 Lima 사용)
# 또는 원격 서버에 마스터 노드 설치

# 원격 서버를 마스터로 사용하는 것을 권장
ssh user@remote-server
curl -sfL https://get.k3s.io | sh -

# 토큰 확인
sudo cat /var/lib/rancher/k3s/server/node-token
```

**워커 노드 (원격 서버):**
```bash
# 다른 원격 서버에서
curl -sfL https://get.k3s.io | K3S_URL=https://<마스터-IP>:6443 \
    K3S_TOKEN=<노드-토큰> sh -
```

**맥북에서 접근:**
```bash
# 마스터 서버에서 kubeconfig 복사
scp user@remote-server:/etc/rancher/k3s/k3s.yaml ~/.kube/config-k3s

# 서버 주소 수정
sed -i '' 's/127.0.0.1/<마스터-IP>/g' ~/.kube/config-k3s

# 컨텍스트 사용
export KUBECONFIG=~/.kube/config-k3s
kubectl get nodes
```

### 옵션 2: Kubeadm (표준 방식)

**마스터 노드 (원격 서버):**
```bash
# 1. kubeadm으로 클러스터 초기화
sudo kubeadm init --pod-network-cidr=10.244.0.0/16

# 2. kubeconfig 설정
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

# 3. CNI 설치 (Flannel)
kubectl apply -f https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml

# 4. 조인 명령어 생성
kubeadm token create --print-join-command
```

**워커 노드 (다른 원격 서버):**
```bash
# 마스터에서 출력된 명령어 실행
sudo kubeadm join <마스터-IP>:6443 --token <token> \
    --discovery-token-ca-cert-hash sha256:<hash>
```

**맥북에서 접근:**
```bash
# 마스터 서버에서 kubeconfig 복사
scp user@remote-server:~/.kube/config ~/.kube/config-remote

# 컨텍스트 사용
export KUBECONFIG=~/.kube/config-remote
kubectl get nodes
```

### 옵션 3: Kind (로컬 멀티 노드)

로컬에서 멀티 노드 클러스터를 시뮬레이션:

```bash
# Kind 설치 (macOS)
brew install kind

# 멀티 노드 클러스터 설정
cat <<EOF > kind-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
- role: worker
EOF

# 클러스터 생성
kind create cluster --config kind-config.yaml --name multi-node

# 확인
kubectl get nodes
```

### 옵션 4: Minikube (멀티 노드 지원)

```bash
# Minikube 설치
brew install minikube

# 멀티 노드 클러스터 생성
minikube start --nodes 3 --driver=docker

# 확인
kubectl get nodes
```

### 옵션 5: 클라우드 K8s (가장 실용적)

실제 원격 서버를 사용하고 싶다면:

**GKE (Google Kubernetes Engine):**
```bash
gcloud container clusters create my-cluster \
    --num-nodes=3 \
    --zone=asia-northeast3-a
```

**EKS (AWS):**
```bash
eksctl create cluster \
    --name my-cluster \
    --region ap-northeast-2 \
    --nodegroup-name workers \
    --nodes 3
```

**Self-Hosted (Kubeadm on VMs):**
- 클라우드 VM 3대 생성 (마스터 1, 워커 2)
- 위의 Kubeadm 방식으로 설치

---

## 비교표

| 방식 | 난이도 | 멀티노드 | 원격 서버 | 권장도 |
|------|--------|----------|-----------|--------|
| Docker Desktop + 원격 조인 | 매우 어려움 | ⚠️ | ✅ | ❌ |
| K3s | 쉬움 | ✅ | ✅ | ⭐⭐⭐⭐⭐ |
| Kubeadm | 보통 | ✅ | ✅ | ⭐⭐⭐⭐ |
| Kind | 쉬움 | ✅ | ❌ (로컬) | ⭐⭐⭐ |
| Minikube | 쉬움 | ✅ | ❌ (로컬) | ⭐⭐⭐ |
| 클라우드 K8s | 매우 쉬움 | ✅ | ✅ | ⭐⭐⭐⭐⭐ |

---

## 추천 시나리오

### 개발/테스트 환경
**상황:** 로컬에서 멀티 노드 테스트
**추천:** Kind 또는 Minikube
```bash
kind create cluster --config kind-multi-node.yaml
```

### 학습/실험 환경
**상황:** 실제 원격 서버 사용 학습
**추천:** K3s (가장 간단)
```bash
# 원격 서버에서
curl -sfL https://get.k3s.io | sh -
```

### 프로덕션 환경
**상황:** 실제 서비스 배포
**추천:** Kubeadm 또는 클라우드 K8s
- Kubeadm: 완전한 제어, 자체 서버
- 클라우드: 관리 부담 감소, 고가용성

### 현재 상황 (원격 서버 조인 원함)
**추천:** K3s
1. 원격 서버에 K3s 마스터 설치
2. 추가 원격 서버를 워커로 조인
3. 맥북에서 kubectl로 접근

```bash
# 1. 원격 서버 (10.0.1.10)에 마스터 설치
ssh user@10.0.1.10
curl -sfL https://get.k3s.io | sh -
sudo cat /var/lib/rancher/k3s/server/node-token

# 2. 다른 원격 서버 (10.0.1.20)를 워커로 추가
ssh user@10.0.1.20
curl -sfL https://get.k3s.io | K3S_URL=https://10.0.1.10:6443 \
    K3S_TOKEN=<토큰> sh -

# 3. 맥북에서 접근
scp user@10.0.1.10:/etc/rancher/k3s/k3s.yaml ~/.kube/config-k3s
sed -i '' 's/127.0.0.1/10.0.1.10/g' ~/.kube/config-k3s
export KUBECONFIG=~/.kube/config-k3s
kubectl get nodes
```

---

## 결론

**Docker Desktop Kubernetes에 원격 노드 추가는 권장하지 않습니다.**

대신:
1. **빠른 시작**: K3s 사용 (원격 서버에 마스터 설치)
2. **표준 방식**: Kubeadm 사용
3. **로컬 테스트**: Kind 또는 Minikube 사용
4. **프로덕션**: 클라우드 K8s 또는 자체 Kubeadm 클러스터

도움이 필요하시면 말씀해주세요!
