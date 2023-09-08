#!/bin/bash

set -o errexit

KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-kind}"
KIND_CLUSTER_OPTS="--name ${KIND_CLUSTER_NAME}"

cat <<EOF | kind create cluster ${KIND_CLUSTER_OPTS} --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 30080
    protocol: TCP
EOF

# install nginx-ingress
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

namespace="ingress-nginx"
label_selector="app.kubernetes.io/component=controller"
timeout=90

# Wait for the resource to exist
while ! kubectl get pods -n $namespace -l $label_selector --no-headers | grep -q "Running"; do
  echo "Waiting for the Pod to be created... up to 60s or so"
  sleep 5
done

echo "Pod is created, waiting for it to be ready..."
# Now wait for the resource to be ready
kubectl wait --namespace $namespace \
  --for=condition=ready pod \
  --selector=$label_selector \
  --timeout=${timeout}s