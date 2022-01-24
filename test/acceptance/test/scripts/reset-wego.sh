#!/bin/bash

ns=$1

function removeResources {

  local ns=$1
  local singleResourceName=$2
  local pluralResourceName=$3

  echo "Deleting ${pluralResourceName}";
  local resources=$(kubectl get $pluralResourceName -o jsonpath='{range .items[*]}{@.metadata.name}{"\n"}' -n $ns)
  echo "$resources" | while IFS= read -r resource ;
  do
    echo " Deleting resource ${resource}";
    local output=$(kubectl delete -n $ns $singleResourceName/${resource} --timeout=20s 2>&1)
    if [[ $output == *"timed out waiting for the condition"* ]]; then
      kubectl patch $singleResourceName/${resource} -n $ns -p '{"metadata":{"finalizers":[]}}' --type=merge
    fi
  done
}

removeResources $ns helmchart helmcharts
removeResources $ns kustomization kustomizations
removeResources $ns gitrepository gitrepositories
removeResources $ns helmrelease helmreleases
removeResources $ns helmchart helmcharts
removeResources $ns helmrepository helmrepositories

echo "Delete any running applications"
kubectl delete apps -n $ns --all
echo "Delete all secrets"
for s in $(kubectl get secrets -n $ns| grep weave-gitops-|cut -d' ' -f1); do kubectl delete secrets $s -n $ns; done

echo "Deleting resources from all my-nginx-* namespaces..."
namespaces="$(kubectl get ns -A --no-headers -o custom-columns=":metadata.name" | grep nginx)";
for ns in $namespaces
do
    kubectl delete all --all -n $ns
    echo "All resources have been deleted from namespace $ns, proceeding to delete empty namespace..."
    kubectl delete ns $ns
done
echo "Test Script run successful :)"