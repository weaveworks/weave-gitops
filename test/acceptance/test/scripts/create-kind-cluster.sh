#!/bin/bash

kind delete clusters --all
kind create cluster --name=$1 --image=$2 --config=./configs/kind-config.yaml