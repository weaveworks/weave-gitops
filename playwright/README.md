# Playwright Tests

### How to run tests locally:

This is a guide to quickly setup the environment to run and debug tests locally on a kind cluster.

There are some prerequisites before running tests locally. It includes installing required tools and environment configurations.

  ## Tools  & Utilities

It is recommended to install latest and stable version of these tools. All tools must be on path.
| Tool | Purpose | Installation |
|--|--|--|
| Docker | Containers runtime environment | `https://docs.docker.com/get-docker` |
| Kind | Running local Kubernetes cluster | `https://kind.sigs.k8s.io/docs/user/quick-start#installation` |
|Kubectl|Kubernetes command-line tool| `https://kubernetes.io/docs/tasks/tools/install-kubectl-linux` |
| Playwright |  a framework for Web Testing and Automation | `https://playwright.dev/docs/intro#installing-playwright`|
| flux | Command-line interface to bootstrap and interact with Flux | `https://fluxcd.io/docs/installation/#install-the-flux-cli`|
| Playwright chromium browser | a browser binary which playwright needs to operate and run tests | After installing Playwright run `playwright install chromium`<br> you can also check this page for more info. <br> `ghttps://playwright.dev/docs/browsers`
| Pytest | a testing framework that allows users to write test codes using Python programming language.  | `https://docs.pytest.org/en/7.1.x/getting-started.html` |
| pytest-reporter-html1 | A basic HTML report for pytest using Jinja2 template engine.   | `https://pypi.org/project/pytest-reporter-html1/` |

## Environment Setup
1. Clone the repo<br/>
    ```bash
    git clone git@github.com:weaveworks/playwright-tests.git
    ```
   
2. Open it in any IDE like **PyCharm** or **VS Code**<p>&nbsp;</p>

3. Launch **Docker Desktop** , for help check this URL [https://docs.docker.com/desktop/install/ubuntu/#launch-docker-desktop](https://docs.docker.com/desktop/install/ubuntu/#launch-docker-desktop) <p>&nbsp;</p>

4. Delete any existing kind cluster(s).
    ```bash
    kind delete clusters --all
    ```
   
5. Create a new clean kind cluster.
    ```bash
    ./utils/scripts/mgmt-cluster-setup.sh kind  $(pwd) playwright-mgmt-kind
    ```
   
6. Make sure that the cluster has been created.
    ```bash
    kind get clusters
    ```
   
7. Setup core and enterprise controllers.
    ```bash
    kubectl create namespace flux-system
    flux install
    kubectl create secret generic git-provider-credentials -n flux-system --from-literal=username="$GITHUB_USER" --from-literal=password="$GITHUB_TOKEN"
    sed -i 's/BRANCH_NAME/${{ steps.extract_branch.outputs.branch_name }}/' ./utils/scripts/resources/flux-system-gitrepo.yaml
    ./utils/scripts/wego-enterprise.sh setup ./utils/scripts
    ```
   
8. Install violating-app.
    ```bash
    kubectl apply -f  ./utils/data/violating-podinfo-kustomization.yaml
    ```
   
9. Install policies.
    ```bash
    kubectl apply -f  ./utils/data/policies.yaml
    ```

10. Flux reconcile violating app.
    ```bash
    flux reconcile kustomization violating-podinfo -n default --with-source || true
    kubectl get pods -A
    ```
    
11. Install gitopsset.
    ```bash
    kubectl apply -f  ./utils/data/gitops-sets-kustomization.yaml
    ```
   
## Run Tests

`export URL="http://localhost:8000"`

`export USER_NAME=""`  -------> you can get it from [./utils/scripts/resources/cluster-user-auth.yaml](./utils/scripts/resources/cluster-user-auth.yaml)

`export PASSWORD=""`  --------> you can get it from [./utils/scripts/resources/cluster-user-auth.yaml](./utils/scripts/resources/cluster-user-auth.yaml)

`export PYTHONPATH=./`

`pytest -s -v --template=html1/index.html --report=test-results/report.html`

## Check the test run report
After running the tests using GitHub Actions, just open the **workflow Summary** page and you will see :
1. a section called **Tests** which contains a table has **Total** number of run tests, many tests with status **Passed**, how many tests with status **Failed** and how many **Skipped** tests.
2. In case there are **Failed** tests you will see in the **failed** section the names of the failed tests with detailed error messages for each test.

## Test Artifacts 
It is just a compressed folder produced during runtime,all you need just open the **workflow Summary** page and download it to your machine and extract it then you will find that it contains **3** reports :
1. **test-run-report.html** which is an HTML report displays a graph for **Total** number of running test cases, how many **Passed** tests and how many **Failed** in addition to a List of the **executed tests** by **name** and the **status** for each one. **To open it just double-click**.This is how it looks like :point_down:

![screencapture-file-home-taghreed-Desktop-report-html-2023-11-21-16_45_44](https://github.com/weaveworks/playwright-tests/assets/44777049/7d882812-c7c3-4390-9df9-a6ea74943a37)

2. **junit_test_report.xml** which is an XML report displays a List of the **executed tests** by **name** and the **status** for each one in addition to the **Failure errors** for the **Failed** tests. **To open it just double-click**.

3. **execution-tracing.zip** which is a compressed folder contains **recorded Playwright traces** after the tests have been run. Traces are a great way for **debugging your tests when they fail on CI**. You can **open traces locally or in your browser on** [trace.playwright.dev](https://trace.playwright.dev/).This is how it looks like :point_down:

![trace viewer](https://github.com/weaveworks/playwright-tests/assets/44777049/dd374fc9-f7d8-4ea1-b1c7-0360822010b6)

**For more information about Playwright Trace Viewer check this URL (https://playwright.dev/docs/trace-viewer)**
 
