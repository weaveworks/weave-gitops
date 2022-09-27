// This may help you refresh your screenshots - but most likely it'll
// be moody and not work.
const puppeteer = require("puppeteer");
const { execSync, spawn, spawnSync } = require("child_process");

const k8sVersion = "1.25.0";
const kindName = "gitops-screenshots";
const fluxVersion = "0.34.0";
const githubUser = process.env.GITHUB_USER;
if (!githubUser) {
  console.log("You need to export a GITHUB_USER");
  process.exit(1);
}

const gitopsTag = process.argv[2];
if (!gitopsTag) {
  console.log("You need to provide the gitops tag you want");
  process.exit(1);
}

function execPrint(command) {
  return execSync(command, { stdio: "inherit" });
}

execPrint(`kind delete cluster --name=${kindName} || :`);
execPrint(
  `kind create cluster --name=${kindName} --image=kindest/node:v${k8sVersion} || :`
);
execPrint(`kubectl config set-context kind-${kindName}`);
execPrint(`rm ./flux || :`);
execPrint(
  `curl -L https://github.com/fluxcd/flux2/releases/download/v${fluxVersion}/flux_${fluxVersion}_linux_amd64.tar.gz | tar xzf -`
);
execPrint(
  `./flux bootstrap github --owner=${githubUser} --repository=fleet-infra --branch=main --path=./clusters/screenshots-cluster --personal`
);
execPrint(`kubectl apply -f - <<EOF
---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  annotations:
    metadata.weave.works/description: This is the Weave GitOps Dashboard.  It provides
      a simple way to get insights into your GitOps workloads.
  labels:
    app.kubernetes.io/component: ui
    app.kubernetes.io/created-by: weave-gitops-cli
    app.kubernetes.io/name: weave-gitops-dashboard
    app.kubernetes.io/part-of: weave-gitops
  name: ww-gitops
  namespace: flux-system
spec:
  interval: 1h0m0s
  url: https://helm.gitops.weave.works
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: ww-gitops
  namespace: flux-system
spec:
  chart:
    spec:
      chart: weave-gitops
      sourceRef:
        kind: HelmRepository
        name: ww-gitops
  interval: 1h0m0s
  values:
    image:
      tag: ${gitopsTag}
    adminUser:
      create: true
      passwordHash: \\$2a\\$10\\$jWoc1KU330tgyeXYgIPmCO/32AXG.hmLvB2UTgIlhRw3wT1kUlruq
      username: admin
EOF`);

(async () => {
  let i = 0;
  while (i < 300) {
    const result = spawnSync(
      "kubectl get pod -n flux-system | grep ww-gitops | grep -q Running",
      { shell: true }
    );
    if (result.status === 0) {
      break;
    }
    await new Promise((r) => setTimeout(r, 1000));
    i += 1;
  }
  childProcess = spawn(
    "kubectl",
    [
      "port-forward",
      "-n",
      "flux-system",
      "svc/ww-gitops-weave-gitops",
      "9001:9001",
    ],
    { detached: true, stdio: "inherit" }
  );

  const browser = await puppeteer.launch();
  const page = await browser.newPage();
  page.setViewport({ width: 1920, height: 1088 });

  i = 0;
  while (i < 300) {
    try {
      await page.goto("http://localhost:9001");
      await page.waitForSelector("#email");
      break;
    } catch (e) {
      console.log(e);
      await new Promise((r) => setTimeout(r, 1000));
      i += 1;
    }
  }
  try {
    await page.waitForTimeout(500); // It takes time for the blue background to show up
    await page.screenshot({ path: "static/img/dashboard-login.png" });

    await page.type("#email", "admin");
    await page.type("#password", "dev");
    const buttons = await page.$$("button");
    buttons[buttons.length - 1].click();

    await page.waitForSelector("[role=tablist]");
    await page.waitForSelector('[aria-label="simple table"]');
    await page.waitForTimeout(500);
    await page.screenshot({
      path: "static/img/dashboard-applications-overview.png",
    });

    await page.click('a[href="/sources"]');
    await page.waitForSelector('a[href="/sources"].selected');
    await page.waitForSelector('[aria-label="simple table"]');
    await page.waitForTimeout(500);
    await page.screenshot({ path: "static/img/dashboard-sources.png" });

    await page.click('a[href="/flux_runtime"]');
    await page.waitForSelector('a[href="/flux_runtime"].selected');
    await page.waitForSelector('[aria-label="simple table"]');
    await page.waitForTimeout(500);
    await page.screenshot({ path: "static/img/dashboard-flux-runtime.png" });

    await page.click('a[href="/flux_runtime/crds"]');
    await page.waitForSelector('a[href="/flux_runtime/crds"].active-tab');
    await page.waitForSelector('[aria-label="simple table"]');
    await page.waitForTimeout(500);
    await page.screenshot({
      path: "static/img/dashboard-flux-runtime-crd.png",
    });

    await page.click('a[href="/applications"]');
    await page.waitForSelector('a[href="/applications"].selected');
    await page.waitForXPath('//tr//td[2]//a[contains(span, "flux-system")]');
    let link = await page.$x('//tr//td[2]//a[contains(span, "flux-system")]');
    await link[0].click();

    await page.waitForXPath('//a[contains(span, "GitRepository")]');
    await page.waitForTimeout(500);
    await page.screenshot({
      path: "static/img/dashboard-application-flux.png",
    });

    link = await page.$x('//a[contains(span, "Events")]');
    await link[0].click();
    await page.waitForTimeout(500);
    await page.screenshot({
      path: "static/img/dashboard-application-events.png",
    });

    link = await page.$x('//a[contains(span, "Graph")]');
    await link[0].click();
    await page.waitForTimeout(500);
    await page.screenshot({
      path: "static/img/dashboard-application-reconciliation.png",
    });

    link = await page.$x('//a[contains(span, "Yaml")]');
    await link[0].click();
    await page.waitForTimeout(500);
    await page.screenshot({
      path: "static/img/dashboard-application-yaml.png",
    });

    await page.click('a[href="/sources"]');
    await page.waitForSelector('a[href="/sources"].selected');
    await page.waitForXPath('//tr//td[2]//a[contains(span, "flux-system")]');
    link = await page.$x('//tr//td[2]//a[contains(span, "flux-system")]');
    await link[0].click();

    await page.waitForXPath('//td[contains(span, "Ref:")]');
    await page.waitForTimeout(500);
    await page.screenshot({ path: "static/img/dashboard-source-flux.png" });

    execPrint(
      `flux create source git podinfo --url=https://github.com/stefanprodan/podinfo --branch=master --interval=30s`
    );
    execPrint(`(flux delete kustomization podinfo -s && sleep 5) || :`);
    execPrint(
      `flux create kustomization podinfo --target-namespace=flux-system --source=podinfo --path="./kustomize" --prune=true --interval=5m`
    );

    await page.click('a[href="/applications"]');
    await page.waitForSelector('a[href="/applications"].selected');
    await page.waitForXPath('//tr//td[2]//a[contains(span, "flux-system")]');
    await page.waitForTimeout(500);
    await page.screenshot({
      path: "static/img/dashboard-applications-with-podinfo.png",
    });

    link = await page.$x('//tr//td[2]//a[contains(span, "podinfo")]');
    await link[0].click();
    await page.waitForXPath('//a[contains(span, "GitRepository")]');
    await page.waitForTimeout(500);
    await page.screenshot({ path: "static/img/dashboard-podinfo-details.png" });

    execPrint(`kubectl apply -f - <<EOF
---
apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
kind: Kustomization
metadata:
  name: podinfo
  namespace: flux-system
spec:
  interval: 60m0s
  path: ./kustomize
  prune: true
  sourceRef:
    kind: GitRepository
    name: podinfo
  targetNamespace: flux-system
  patches:
    - patch: |-
        apiVersion: autoscaling/v2beta2
        kind: HorizontalPodAutoscaler
        metadata:
          name: podinfo
        spec:
          minReplicas: 3
      target:
        name: podinfo
        kind: HorizontalPodAutoscaler
EOF`);
    execPrint(`flux reconcile kustomization podinfo`);
    await page.waitForXPath(
      '//table[count(//tr//td[2][contains(span, "Pod")]) = 4]'
    ); // HorizontalPodAutoscaler causes off-by-one
    await page.waitForTimeout(500);
    await page.screenshot({ path: "static/img/dashboard-podinfo-updated.png" });

    link = await page.$x('//button[contains(span, "Suspend")]');
    await link[0].click();
    await page.waitForTimeout(5500); // Wait for banner to go away
    await page.screenshot({
      path: "static/img/dashboard-podinfo-details-suspended.png",
    });

    await page.click('a[href="/applications"]');
    await page.waitForSelector('a[href="/applications"].selected');
    await page.waitForXPath('//tr//td[2]//a[contains(span, "flux-system")]');
    await page.waitForTimeout(500);
    await page.screenshot({
      path: "static/img/dashboard-podinfo-suspended.png",
    });

    link = await page.$x('//tr//td[2]//a[contains(span, "podinfo")]');
    await link[0].click();
    await page.waitForTimeout(500);
    link = await page.$x('//button[contains(span, "Resume")]');
    await link[0].click();
    await page.waitForTimeout(5500); // Wait for banner to go away
    await page.screenshot({ path: "static/img/dashboard-podinfo-updated.png" });
  } catch (e) {
    await page.screenshot({ path: "error.png" });
    await browser.close();
    childProcess.kill();

    throw e;
  }

  await browser.close();
  childProcess.kill();
})();
