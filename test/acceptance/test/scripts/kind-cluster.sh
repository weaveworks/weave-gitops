#!/bin/bash

echo "Delete existing kind clusters"
kind delete clusters --all
echo "Create a new kind cluster with name "$1
kind create cluster --name=$1 --image=$2 --config=./configs/kind-config.yaml
echo "Wait for pods to be ready "$1
kubectl wait --for=condition=Ready --timeout=300s -n kube-system --all pod