#!/bin/bash

echo "Create a new kind cluster with name "$1
kind create cluster --name=$1 --image=$2 --config=./configs/kind-config.yaml
