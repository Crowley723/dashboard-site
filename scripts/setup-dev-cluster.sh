#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

CLUSTER_NAME="conduit-dev"

echo -e "${BLUE}Creating k3d cluster...${NC}"
if k3d cluster list | grep -q "^${CLUSTER_NAME}"; then
  echo -e "${YELLOW}Cluster '${CLUSTER_NAME}' already exists.${NC}"
  echo -e "${YELLOW}Delete it first with: make dev-cluster-delete${NC}"
  exit 1
fi

k3d cluster create ${CLUSTER_NAME} \
  --port 9080:30080@server:0 \
  --port 9443:30443@server:0 \
  --api-port 6550 \
  --servers 1 \
  --agents 0

echo -e "${BLUE}Installing cert-manager...${NC}"
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml

echo -e "${BLUE}Waiting for cert-manager to be ready...${NC}"
kubectl wait --for=condition=Available --timeout=300s \
  deployment/cert-manager -n cert-manager
kubectl wait --for=condition=Available --timeout=300s \
  deployment/cert-manager-webhook -n cert-manager
kubectl wait --for=condition=Available --timeout=300s \
  deployment/cert-manager-cainjector -n cert-manager

echo -e "${BLUE}Creating self-signed ClusterIssuer...${NC}"
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
EOF

echo -e "${BLUE}Creating namespaces...${NC}"
kubectl create namespace conduit || true

echo -e "${GREEN}Dev cluster ready!${NC}"
echo ""
echo "Cluster info:"
kubectl cluster-info --context k3d-${CLUSTER_NAME}
echo ""
echo "Next steps:"
echo "  - Run your app: make dev-local"
echo "  - Check status: make dev-cluster-status"
echo "  - Delete cluster: make dev-cluster-delete"