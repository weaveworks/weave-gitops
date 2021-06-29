#!/bin/bash

echo "Delete all kustomizations"
kubectl delete -n $1 kustomizations.kustomize.toolkit.fluxcd.io --all
echo "Delete all gitrepositories"
kubectl delete -n $1 gitrepositories.source.toolkit.fluxcd.io --all
echo "Delete all secrets"
for s in $(kubectl get secrets -n $1| grep weave-gitops-|cut -d' ' -f1); do kubectl delete secrets $s -n $1; done