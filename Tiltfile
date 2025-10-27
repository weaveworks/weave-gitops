allow_k8s_contexts('kind-wego-dev')

if os.getenv('MANUAL_MODE'):
   trigger_mode(TRIGGER_MODE_MANUAL)

image_repository = os.getenv('IMAGE_REPO', 'localhost:5001/weaveworks/wego-app')

load('ext://restart_process', 'docker_build_with_restart')

advanced_go_dev_mode = os.getenv('FAST_AND_FURIOUSER')
skip_ui_build = os.getenv("SKIP_UI_BUILD")

local_resource(
    'install-flux',
    '''
    kubectl wait --for=condition=Ready nodes --all --timeout=300s || exit 1
    
    echo "Verifying cluster access..."
    kubectl cluster-info || exit 1
    kubectl get nodes || exit 1
    kubectl apply -f ./tools/dev-resources/namespace.yaml
    

    flux install --components-extra="image-reflector-controller,image-automation-controller" || exit 1
    
    kubectl get namespace flux-system || exit 1
    
    kubectl wait --for=condition=Ready pod \
        -l app=source-controller \
        -l app=kustomize-controller \
        -l app=helm-controller \
        -l app=image-reflector-controller \
        -l app=image-automation-controller \
        -n flux-system \
        --timeout=120s || exit 1
        
    echo "Flux installation complete"
    ''',
    labels=['setup'],
    allow_parallel=False,
)

if advanced_go_dev_mode:

    local_resource(
        'gitops-server',
        'GOOS=linux make gitops-server',
        deps=[
            './cmd',
            './pkg',
            './core',
            './api',
        ]
    )

    if not skip_ui_build:
        local_resource(
            'ui-server',
            'make ui',
            deps=[
                './ui',
            ]
        )

    docker_build_with_restart(
        image_repository,
        '.',
        only=[
            './bin',
        ],
        dockerfile="dev.dockerfile",
        entrypoint="/app/build/gitops-server --log-level=debug --insecure",
        live_update=[
            sync('./bin', '/app/build'),
        ],
    )
else:
    docker_build(
        image_repository,
        '.',
        dockerfile="gitops-server.dockerfile",
    )

k8s_yaml(helm('./charts/gitops-server', name='dev', values='./tools/helm-values-dev.yaml', set=['image.repository=' + image_repository]))
k8s_yaml(helm('./tools/charts/dev', name='dev', values='./tools/charts/dev/values.yaml'))

deps = ['gitops-server'] if advanced_go_dev_mode else []

if advanced_go_dev_mode:
    if not skip_ui_build:
        deps.append('ui-server')

deps.append('install-flux')

k8s_resource('dev-weave-gitops', port_forwards='9001', resource_deps=deps, labels=['app'])
local_resource(
    'apply-dev-resources',
    '''
    sleep 30
    until kubectl wait --for=condition=Established crd gitrepositories.source.toolkit.fluxcd.io && \
          kubectl wait --for=condition=Established crd helmrepositories.source.toolkit.fluxcd.io  && \
          kubectl wait --for=condition=Established crd kustomizations.kustomize.toolkit.fluxcd.io  && \
          kubectl wait --for=condition=Established crd helmreleases.helm.toolkit.fluxcd.io  && \
          kubectl wait --for=condition=Established crd imagerepositories.image.toolkit.fluxcd.io  && \
          kubectl wait --for=condition=Established crd imagepolicies.image.toolkit.fluxcd.io  && \
          kubectl wait --for=condition=Established crd imageupdateautomations.image.toolkit.fluxcd.io; do
        sleep 5
    done
    
    
    kubectl apply -k ./tools/dev-resources
    sleep 10
    
    until kubectl wait --for=condition=Established crd clusterissuers.cert-manager.io && \
          kubectl wait --for=condition=Established crd certificates.cert-manager.io && \
          kubectl wait --for=condition=Established crd certificaterequests.cert-manager.io && \
          kubectl wait --for=condition=Established crd issuers.cert-manager.io; do
        sleep 5
    done

    kubectl get helmreleases --all-namespaces || true
    
    kubectl wait --for=condition=Ready \
        helmrelease.helm.toolkit.fluxcd.io/cert-manager \
        -n cert-manager \
        --timeout=600s || {
        kubectl describe helmrelease cert-manager -n cert-manager || true
        exit 1
    }
    
    kubectl wait --for=condition=Ready \
        helmrelease.helm.toolkit.fluxcd.io/ingress-nginx \
        -n ingress-nginx \
        --timeout=600s || {
        kubectl describe helmrelease ingress-nginx -n ingress-nginx || true
        exit 1
    }
    
    kubectl wait --for=condition=Ready \
        helmrelease.helm.toolkit.fluxcd.io/kube-prometheus-stack \
        -n monitoring \
        --timeout=600s || {
        kubectl describe helmrelease kube-prometheus-stack -n monitoring || true
        exit 1
    }
    
    kubectl get kustomizations --all-namespaces || true
    
    kubectl wait --for=condition=Ready \
        kustomization.kustomize.toolkit.fluxcd.io/kube-prometheus-stack \
        kustomization.kustomize.toolkit.fluxcd.io/monitoring-config \
        -n monitoring \
        --timeout=600s || {
        kubectl describe kustomizations -n monitoring || true
        exit 1
    }
    
    kubectl wait --for=condition=Ready \
        kustomization.kustomize.toolkit.fluxcd.io/podinfo \
        -n default \
        --timeout=600s || {
        kubectl describe kustomization podinfo -n default || true
        exit 1
    }
    ''',
    resource_deps=['install-flux', 'dev-weave-gitops'],
    labels=['setup'],
    allow_parallel=False,
)

local_resource(
    'certs',
    '''
    kubectl apply -k ./tools/dev-resources/certs
    ''',
    resource_deps=['apply-dev-resources'],
    labels=['setup'],
    allow_parallel=False,
)

local_resource(
    'playwright-tests',
    '''
    export URL="http://localhost:9001"
    export USER_NAME="dev"
    export PASSWORD="dev"
    export CLUSTER_NAME="wego-dev"
    export PYTHONPATH="./playwright"
    
    if [ ! -d "playwright/venv" ]; then
        cd playwright
        python3 -m venv venv
        source venv/bin/activate
        pip install -r requirements.txt || pip install playwright pytest pytest-reporter-html1
        playwright install chromium
        cd ..
    fi
    
    cd playwright
    source venv/bin/activate
    pytest -s -v --template=html1/index.html --report=test-results/report.html
    cd ..
    ''',
    resource_deps=['dev-weave-gitops', 'apply-dev-resources'],
    labels=['test'],
    allow_parallel=False,
)

print("""
╔══════════════════════════════════════════════════════════════════════╗
║                Weave GitOps Local Development                        ║
╚══════════════════════════════════════════════════════════════════════╝

🚀 Services will be available at:

  📦 Weave GitOps:     http://localhost:9001
  📊 Monitoring:       http://localhost:3000 (if deployed)
  
🏗️  Startup Order:
  1. Kind cluster 'wego-dev' (created if needed)
  2. Flux CD (installed with image automation)
  3. Dev resources (cert-manager, ingress, podinfo, monitoring)
  4. Weave GitOps server

🔧 Quick Actions:
  - Tilt will auto-reload on code changes
  - Check cluster: kubectl cluster-info --context kind-wego-dev
  - View Flux status: flux get sources
  - View Podinfo: kubectl get pods -n default

🧪 Testing:
  - Playwright tests run automatically after deployment
  - Manual test: just test-playwright
  - Setup Playwright: just test-playwright-setup

════════════════════════════════════════════════════════════════════════
""")
