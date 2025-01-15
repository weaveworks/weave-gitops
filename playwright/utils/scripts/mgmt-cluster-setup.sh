#!/usr/bin/env bash

args=("$@")

if [ -z ${args[0]} ] || ([ ${args[0]} != 'eks' ] && [ ${args[0]} != 'gke' ] && [ ${args[0]} != 'kind' ])
then
    echo "Invalid option, valid values => [ eks, gke, kind ]"
    exit 1
fi

set -x

function setup_eks {
  if [ ${#args[@]} -ne 4 ]
  then
    echo "All required arguments for EKS cluster setup must be provided => [ WORKSPACE_PATH, CLUSTER_NAME, AWS_REGION ] "
    exit 1
  fi

  export CLUSTER_NAME=${args[2]}
  export AWS_REGION=${args[3]}
  export CLUSTER_VERSION=1.23

  export CLUSTER_EXISTS=$(eksctl get clusters | grep -i $CLUSTER_NAME)
  if [ -z $CLUSTER_EXISTS ]; then
    # Create EKS cluster
    eksctl create cluster --name=$CLUSTER_NAME --version=$CLUSTER_VERSION
  fi

  eksctl utils write-kubeconfig --cluster=$CLUSTER_NAME --kubeconfig=$HOME/.kube/config
  kubectl config rename-context $(kubectl config current-context) $(kubectl config get-clusters | grep $CLUSTER_NAME | sed s/_/-/g)

  kubectl get nodes -A
  kubectl get all --all-namespaces -o wide

  # Associate oidc identity provider to the cluster
  cat ${args[1]}/test/utils/data/oidc/oidc-eks-associate-identity-provider.yaml | \
    sed s,{{CLUSTER_NAME}},${CLUSTER_NAME},g | \
    sed s,{{AWS_REGION}},${AWS_REGION},g | \
    sed s,{{ISSUER_URL}},${OIDC_ISSUER_URL},g | \
    sed s,{{CLIENT_ID}},${DEX_CLI_CLIENT_ID},g | \
    eksctl associate identityprovider -f -
  eksctl utils associate-iam-oidc-provider --cluster ${CLUSTER_NAME} --approve

  # Check if open id connect provided is alreadey associated with the cluster
  ID_PROVIDER=$(aws eks describe-cluster --name $CLUSTER_NAME --query "cluster.identity.oidc.issuer" --output text | cut -c 9-100)
  PROVIDER_EXISTS=$(aws iam list-open-id-connect-providers | grep $ID_PROVIDER)

  if [ -z $PROVIDER_EXISTS ]; then
    # Associating oidc provider takes very long time to update the cluster.
    for i in {0..1}
    do
      id=$(aws eks describe-cluster --name $CLUSTER_NAME --query "cluster.identity.oidc.issuer" --output text | cut -c 9-100)
      provider=$(aws iam list-open-id-connect-providers | grep $id)
      if [ -z $provider ]; then
        sleep 1m
      else
        break
      fi
    done
  fi

  eksctl utils write-kubeconfig --cluster=$CLUSTER_NAME --kubeconfig=${OIDC_KUBECONFIG}

  kubectl --kubeconfig=${OIDC_KUBECONFIG} config set-credentials oidc-user \
    --exec-api-version=client.authentication.k8s.io/v1beta1 \
    --exec-command=kubectl \
    --exec-arg=oidc-login \
    --exec-arg=get-token \
    --exec-arg=--oidc-issuer-url=${OIDC_ISSUER_URL} \
    --exec-arg=--oidc-client-id=${DEX_CLI_CLIENT_ID} \
    --exec-arg=--oidc-client-secret=${DEX_CLI_CLIENT_SECRET} \
    --exec-arg=--oidc-extra-scope="openid email groups offline_access" \
    --exec-arg=--skip-open-browser

  kubectl --kubeconfig=${OIDC_KUBECONFIG} config set-context --current --user=oidc-user

  exit 0
}

function setup_gke {
  if [ ${#args[@]} -ne 4 ]
  then
    echo "All required arguments for GKE cluster setup must be provided => [ WORKSPACE_PATH, CLUSTER_NAME, CLUSTER_REGION ] "
    exit 1
  fi

  export CLUSTER_NAME=${args[2]}
  export CLUSTER_REGION=${args[3]}
  export CLUSTER_VERSION=1.23.13

  export CLUSTER_EXISTS=$(gcloud container clusters list | grep -i $CLUSTER_NAME)
  if [ -z $CLUSTER_EXISTS ]; then
    # Creates GKE cluster
    gcloud container clusters create $CLUSTER_NAME --cluster-version=$CLUSTER_VERSION --zone $CLUSTER_REGION --enable-identity-service
  fi

  gcloud container clusters get-credentials $CLUSTER_NAME --zone $CLUSTER_REGION
  kubectl config rename-context $(kubectl config current-context) $(kubectl config get-clusters | grep $CLUSTER_NAME | sed s/_/-/g)

  kubectl get nodes -A
  kubectl get all --all-namespaces -o wide

  gcloud components install kubectl-oidc

  CA_AUTHORITY=$(kubectl get clientconfig default -n kube-public -o=jsonpath="{.spec.certificateAuthorityData}")
  SERVER_NAME=$(kubectl get clientconfig default -n kube-public -o=jsonpath="{.spec.server}")
  DEX_REDIRECT_URL="http://localhost:8000"

  cat ${args[1]}/test/utils/data/oidc/oidc-gke-client-config.yaml | \
    sed s,{{CA_AUTHORITY}},${CA_AUTHORITY},g | \
    sed s,{{CLUSTER_NAME}},${CLUSTER_NAME},g | \
    sed s,{{SERVER_NAME}},${SERVER_NAME},g | \
    sed s,{{ISSUER_URL}},${OIDC_ISSUER_URL},g | \
    sed s,{{CLIENT_ID}},${DEX_CLI_CLIENT_ID},g | \
    sed s,{{CLIENT_SECRET}},${DEX_CLI_CLIENT_SECRET},g | \
    sed s,{{REDIRECT_URL}},${DEX_REDIRECT_URL},g | \
    kubectl apply -f -

  exit 0
}

function setup_kind {
  if [ ${#args[@]} -ne 3 ]
  then
    echo "All required arguments for Kind cluster setup must be provided => [ WORKSPACE_PATH, CLUSTER_NAME ] "
    exit 1
  fi

  export CLUSTER_NAME=${args[2]}

  kind create cluster --name $CLUSTER_NAME --image=kindest/node:v1.23.4 --config ${args[1]}/utils/data/kind/local-kind-config.yaml
  kubectl wait --for=condition=Ready --timeout=120s -n kube-system pods --all
  kubectl get pods -A
  exit 0
}

if [ ${args[0]} = 'eks' ]; then
    setup_eks
elif [ ${args[0]} = 'gke' ]; then
    setup_gke
elif [ ${args[0]} = 'kind' ]; then
    setup_kind
fi