#!/bin/bash

echo "Delete all kustomizations"
kubectl delete -n $1 kustomizations.kustomize.toolkit.fluxcd.io --all --timeout 30s
echo "Delete all gitrepositories"
kubectl delete -n $1 gitrepositories.source.toolkit.fluxcd.io --all --timeout 30s
echo "Delete all helmrepositories"
kubectl delete -n $1 helmreleases.helm.toolkit.fluxcd.io --all --timeout 30s
kubectl delete -n $1 helmcharts.source.toolkit.fluxcd.io --all --timeout 30s
kubectl delete -n $1 helmrepositories.source.toolkit.fluxcd.io --all --timeout 30s
echo "Delete any running applications"
kubectl delete apps -n $1 --all --timeout 30s
echo "Delete all secrets"
for s in $(kubectl get secrets -n $1| grep weave-gitops-|cut -d' ' -f1); do kubectl delete secrets $s -n $1; done