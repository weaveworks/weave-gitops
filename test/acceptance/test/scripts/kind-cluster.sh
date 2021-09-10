#!/bin/bash

#echo "Delete existing kind clusters"
#kind delete clusters --all
echo "Create a new kind cluster with name "$1
#kind create cluster --name=$1 --image=$2 --config=./configs/kind-config.yaml
#kind create cluster --name=$1 --image=$2 --config=../configs/kind-config.yaml
#kind create cluster --name=$1 --image=$2 --config=test/acceptance/test/configs/kind-config.yaml --wait 5m --kubeconfig
kind create cluster --name=$1 --kubeconfig $2 --image=$3 --config=/Users/joseordaz/go/src/github.com/josecordaz/weave-gitops-ww/test/acceptance/test/configs/kind-config.yaml --wait 5m