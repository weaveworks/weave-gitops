#!/usr/bin/env bash


args=("$@")

if [ -z ${args[0]} ] || ([ ${args[0]} != 'setup' ] && [ ${args[0]} != 'reset' ] && [ ${args[0]} != 'reset_controllers' ])
then 
    echo "Invalid option, valid values => [ setup, reset, reset_controllers ]"
    exit 1
fi

set -x 

function get-external-ip {
  local  __resultvar=$1
  local worker_name
  local external_ip
  
  if [ "$MANAGEMENT_CLUSTER_KIND" == "eks" ] || [ "$MANAGEMENT_CLUSTER_KIND" == "gke" ]; then
    worker_name=$(kubectl get node --selector='!node-role.kubernetes.io/master' -o name | head -n 1 | cut -d '/' -f2-)
    external_ip=$(kubectl get nodes -o jsonpath="{.items[?(@.metadata.name=='${worker_name}')].status.addresses[?(@.type=='ExternalIP')].address}")
  fi
  eval $__resultvar="'$external_ip'"
}

function setup {
  if [ ${#args[@]} -ne 2 ]
  then
    echo "Workspace path is a required argument"
    exit 1
  fi
  
  if [ "$MANAGEMENT_CLUSTER_KIND" == "eks" ] || [ "$MANAGEMENT_CLUSTER_KIND" == "gke" ]; then
    get-external-ip WORKER_NODE_EXTERNAL_IP
    # Configure inbound UI node ports
    if [ "$MANAGEMENT_CLUSTER_KIND" == "eks" ]; then
      INSTANCE_SECURITY_GROUP=$(aws ec2 describe-instances --filter "Name=ip-address,Values=${WORKER_NODE_EXTERNAL_IP}" --query 'Reservations[*].Instances[*].NetworkInterfaces[0].Groups[0].{sg:GroupId}' --output text)
      aws ec2 authorize-security-group-ingress --group-id ${INSTANCE_SECURITY_GROUP}  --ip-permissions FromPort=${UI_NODEPORT},ToPort=${UI_NODEPORT},IpProtocol=tcp,IpRanges='[{CidrIp=0.0.0.0/0}]',Ipv6Ranges='[{CidrIpv6=::/0}]'
    else
      gcloud compute firewall-rules create ui-node-port --allow tcp:${UI_NODEPORT}
      # This allows us to test out cli auth passthrough.
      # Our current system of SelfSubjectAccessReview to determine namespace access
      # does not supported external auth systems like the one GKE configures by default for kubectl etc.
      # We need to add explicit permissions here that will correctly appear in the SelfSubjectAccessReview
      # query made by the clusters-service when responding to get /v1/clusters and /v1/templates etc.
      kubectl apply -f ${args[1]}/test/utils/data/rbac/gke-ci-user-cluster-admin-rolebinding.yaml
    fi
  fi

  # Set enterprise cluster CNAME host entry mapping in the /etc/hosts file
#  ${args[1]}/hostname-to-ip.sh ${MANAGEMENT_CLUSTER_CNAME}
   
  helm repo add wkpv3 https://s3.us-east-1.amazonaws.com/weaveworks-wkp/charts-v3-r2/
  helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
  helm repo add cert-manager https://charts.jetstack.io
  helm repo update  
  
  # Install cert-manager for tls certificate creation
  helm upgrade --install \
    cert-manager cert-manager/cert-manager \
    --namespace cert-manager --create-namespace \
    --version v1.10.0 \
    --wait \
    --set installCRDs=true
  kubectl wait --for=condition=Ready --timeout=120s -n cert-manager --all pod

  # Create admin cluster user secret
  kubectl create secret generic sops-gpg \
  --namespace flux-system \
  --from-literal=sops.asc="${WEAVE_GITOPS_DEV_SOPS_KEY}"

  # Create admin cluster user secret
#  kubectl create secret generic cluster-user-auth \
#  --namespace flux-system \
#  --from-literal=username=wego-admin \
#  --from-literal=password=${CLUSTER_ADMIN_PASSWORD_HASH}

  kubectl apply -f ${args[1]}/resources/cluster-user-auth.yaml

  kubectl apply -f ${args[1]}/resources/flux-system-gitrepo.yaml
  flux reconcile source git -n flux-system flux-system --verbose

  kubectl apply -f ${args[1]}/resources/shared-secrets-kustomization.yaml
  flux reconcile kustomization -n flux-system shared-secrets --verbose

  # Choosing weave-gitops-enterprise chart version to install
  if [ -z ${ENTERPRISE_CHART_VERSION} ]; then
    CHART_VERSION=${DEFAULT_ENTERPRISE_CHART_VERSION}
  else
    CHART_VERSION=${ENTERPRISE_CHART_VERSION}
  fi

  # Install weave gitops enterprise controllers
  helmArgs=()
  helmArgs+=( --set "service.ports.https=8000" )
  helmArgs+=( --set "service.targetPorts.https=8000" )
  helmArgs+=( --set "tls.enabled=false" )
  helmArgs+=( --set "config.oidc.enabled=false" )
  helmArgs+=( --set "policy-agent.enabled=true" )

  helm upgrade --install my-mccp wkpv3/mccp --version "${CHART_VERSION}" --namespace flux-system ${helmArgs[@]} --wait

  # Install ingress-nginx for tls termination 
  command="helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
            --namespace ingress-nginx --create-namespace \
            --version 4.4.0 \
            --wait \
            --set controller.service.type=NodePort \
            --set controller.service.nodePorts.https=30080 \
            --set controller.extraArgs.v=4"  
  # When policy-agent ValidatingWebhook service has not fully started up, admission controller call to matching validating webhook fails.
  # Retrying few times gives enough time for ValidatingWebhook service to become available
  for i in {0..5}
  do
    echo "Attempt installing ingress-nginx: $(($i+1))"
    eval $command
    if [ $? -ne 0 ]; then
    sleep 3
    else          
      break    
    fi  
  done  
  kubectl wait --for=condition=Ready --timeout=120s -n ingress-nginx --all pod
  
  cat ${args[1]}/resources/ingress/certificate-issuer.yaml | \
      sed s,{{HOST_NAME}},"${MANAGEMENT_CLUSTER_CNAME}",g | \
      kubectl apply -f -
  kubectl wait --for=condition=Ready --timeout=60s -n flux-system --all certificate

  cat ${args[1]}/resources/ingress/ingress.yaml | \
      sed s,{{HOST_NAME}},${MANAGEMENT_CLUSTER_CNAME},g | \
      kubectl apply -f -

  # Create profiles HelmReposiotry 'weaveworks-charts'
  flux create source helm weaveworks-charts --url="https://raw.githubusercontent.com/weaveworks/profiles-catalog/gh-pages" --interval=30s --namespace flux-system 

  # Install RBAC for user authentication
  kubectl apply -f ${args[1]}/resources/rbac/user-role-bindings.yaml

  kubectl port-forward -n flux-system svc/clusters-service 8000:8000 &

  kubectl get pods -A

  exit 0
}

function reset {
   kubectl delete ValidatingWebhookConfiguration policy-agent
  # Delete any orphan resources
  kubectl delete CAPITemplate --all
  kubectl delete GitOpsTemplate --all
  kubectl delete ClusterBootstrapConfig --all
  kubectl delete ClusterResourceSet --all
  kubectl delete ClusterRoleBinding clusters-service-impersonator
  kubectl delete ClusterRole clusters-service-impersonator-role 
  kubectl delete crd capitemplates.capi.weave.works clusterbootstrapconfigs.capi.weave.works gitopstemplates.templates.weave.works 
  # Delete flux system from the management cluster
  flux uninstall --silent
  # Delete capi provider
  if [ "$CAPI_PROVIDER" == "capa" ]; then
    clusterctl delete --infrastructure aws
  elif [ "$CAPI_PROVIDER" == "capg" ]; then
    clusterctl delete --infrastructure gcp
  else
    clusterctl delete --infrastructure docker    
  fi
}

function reset_controllers {
    if [ ${#args[@]} -ne 2 ]; then
      echo "Cotroller's type is a required argument, valid values => [ enterprise, core, all ]"
      exit 1
    fi

    
    controllerNames=()
    if [ ${args[1]} == "enterprise" ] || [ ${args[1]} == "all" ]; then
      # Sometime due to the test conditions the cluster service pod is in transition state i.e. one terminating and the new one is being created at the same time.
      # Under such state we have two cluster srvice pods momentarily 
      counter=10
      while [ $counter -gt 0 ]
      do
          CLUSTER_SERVICE_POD=$(kubectl get pods -n flux-system|grep cluster-service|tr -s ' '|cut -f1 -d ' ')
          pod_count=$(echo $CLUSTER_SERVICE_POD | wc -w |awk '{print $1}')
          if [ $pod_count -gt 1 ]
          then            
              sleep 2
              counter=$(( $counter - 1 ))
          else
              break
          fi        
      done
      controllerNames+=" ${CLUSTER_SERVICE_POD}"
    fi

    if [ ${args[1]} == "core" ] || [ ${args[1]} == "all" ]; then
      KUSTOMIZE_POD=$(kubectl get pods -n flux-system|grep kustomize-controller|tr -s ' '|cut -f1 -d ' ')
      controllerNames+=" ${KUSTOMIZE_POD}"
    fi

    kubectl delete -n flux-system pod $controllerNames
}

if [ ${args[0]} = 'setup' ]; then
    setup
elif [ ${args[0]} = 'reset' ]; then
    reset
elif [ ${args[0]} = 'reset_controllers' ]; then
    reset_controllers
fi

