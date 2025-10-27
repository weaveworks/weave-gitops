# Weave GitOps Development Commands

# Start the development environment with Tilt
dev-up:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "🚀 Starting Weave GitOps development environment..."
    echo ""
    
    # Start Docker registry for local project images
    echo "Creating Docker registry..."
    docker network create kind 2>/dev/null || true
    if ! docker inspect kind-registry >/dev/null 2>&1; then
        docker run -d --restart=always \
            -p "127.0.0.1:5001:5000" \
            --network kind \
            --name kind-registry \
            registry:2
        echo "✅ Docker registry started on localhost:5001"
    else
        echo "✅ Docker registry already running"
    fi
    echo ""
    
    # Create kind cluster if it doesn't exist
    if ! kind get clusters 2>/dev/null | grep -q '^wego-dev$'; then
        echo "Creating kind cluster..."
        kind create cluster --config tools/kind/wego-dev-config.yaml
        echo "✅ Kind cluster created"
    else
        echo "✅ Kind cluster already exists"
    fi
    echo ""
    
    # Export kubeconfig to make cluster available to Tilt immediately
    kind export kubeconfig --name wego-dev
    echo "✅ Kubeconfig exported"
    echo ""
    
    # Stop any existing Tilt instance
    echo "Stopping any existing Tilt instances..."
    tilt down 2>/dev/null || true
    pkill -f "tilt" 2>/dev/null || true
    sleep 2
    echo ""
    
    # Start Tilt
    echo "Starting Tilt (press 'space' to open web UI)..."
    tilt up

# Tear down the kind cluster and clean up
dev-down:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "🧹 Tearing down Weave GitOps development environment..."
    echo ""
    
    echo "Stopping Tilt..."
    tilt down || true
    echo ""
    
    echo "Deleting kind cluster..."
    kind delete cluster --name wego-dev || true
    echo ""
    
    echo "Stopping Docker registry..."
    docker stop kind-registry 2>/dev/null || true
    docker rm kind-registry 2>/dev/null || true
    echo ""
    
    echo "✅ Development environment torn down"

# ============================================================================
# PLAYWRIGHT TESTING
# ============================================================================

# Run Playwright tests against the dev environment
test-playwright:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "🧪 Running Playwright tests against dev environment..."
    
    # Check if dev environment is running
    if ! pgrep -f "tilt" > /dev/null; then
        echo "❌ Tilt is not running. Start with: just dev-up"
        exit 1
    fi
    
    # Set up environment variables for tests
    export URL="http://localhost:9001"
    export USER_NAME="dev"
    export PASSWORD="dev"
    export CLUSTER_NAME="wego-dev"
    export PYTHONPATH="./playwright"
    
    # Install Python dependencies if needed
    if [ ! -d "playwright/venv" ]; then
        echo "Setting up Python virtual environment..."
        cd playwright
        python3 -m venv venv
        source venv/bin/activate
        pip install playwright pytest pytest-reporter-html1
        playwright install chromium
        cd ..
    fi
    
    # Run tests
    cd playwright
    source venv/bin/activate
    pytest -s -v --template=html1/index.html --report=test-results/report.html
    cd ..

    echo "📊 Test report: playwright/test-results/report.html"

# Setup Playwright test environment (one-time setup)
test-playwright-setup:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "🔧 Setting up Playwright test environment..."
    
    cd playwright
    python3 -m venv venv
    source venv/bin/activate
    pip install playwright pytest pytest-reporter-html1
    playwright install chromium
    cd ..
    
    echo "✅ Playwright environment setup complete"
