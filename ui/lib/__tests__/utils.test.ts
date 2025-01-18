import { JSDOM } from "jsdom";
import { GetVersionResponse } from "../api/core/core.pb";
import { Kind } from "../api/core/types.pb";
import { Automation, HelmRelease, Kustomization } from "../objects";
import {
  convertGitURLToGitProvider,
  convertImage,
  createYamlCommand,
  formatLogTimestamp,
  formatMetadataKey,
  getAppVersion,
  getSourceRefForAutomation,
  isAllowedLink,
  isHTTP,
  makeImageString,
  pageTitleWithAppName,
  statusSortHelper,
  getBasePath,
} from "../utils";

describe("utils lib", () => {
  describe("isHTTP", () => {
    it("detects HTTP", () => {
      expect(isHTTP("http://www.google.com")).toEqual(true);
      expect(isHTTP("https://www.google.com/")).toEqual(true);
      expect(isHTTP("http://10.0.0.1/")).toEqual(true);
      expect(isHTTP("http://127.0.0.1/")).toEqual(true);
      expect(isHTTP("https://192.168.0.1/")).toEqual(true);
      expect(isHTTP("http://192.168.0.2/")).toEqual(true);
      expect(isHTTP("https://169.254.0.1/")).toEqual(true);
      expect(isHTTP("http://169.254.0.2/")).toEqual(true);
      expect(isHTTP("https://172.31.0.1/")).toEqual(true);
      expect(isHTTP("http://172.31.0.2/")).toEqual(true);
      expect(
        isHTTP(
          "http://localhost:8080/applications/argocd/fsa-installation?view=tree",
        ),
      ).toEqual(true);
      expect(
        isHTTP(
          "https://localhost:8080/applications/argocd/fsa-installation?view=tree",
        ),
      ).toEqual(true);
      expect(
        isHTTP("http://github.com/weaveworks/weave-gitops-clusters"),
      ).toEqual(true);
      expect(
        isHTTP("http://github.com/weaveworks/weave-gitops-clusters/"),
      ).toEqual(true);
    });
    it("detects HTTPS", () => {
      expect(isHTTP("https://www.google.com")).toEqual(true);
      expect(isHTTP("https://www.google.com/")).toEqual(true);
      expect(
        isHTTP("https://github.com/weaveworks/weave-gitops-clusters"),
      ).toEqual(true);
      expect(
        isHTTP("https://github.com/weaveworks/weave-gitops-clusters/"),
      ).toEqual(true);
    });
    it("detects non-HTTP string", () => {
      expect(isHTTP("test string")).toEqual(false);
      expect(isHTTP("smtp://server/")).toEqual(false);
      expect(isHTTP("smtp://http/")).toEqual(false);
      expect(isHTTP("smtp://https/")).toEqual(false);
      expect(isHTTP("this is a random http sentence")).toEqual(false);
      expect(isHTTP("this is a random https sentence")).toEqual(false);
      expect(isHTTP("http:// this is a random http sentence")).toEqual(false);
      expect(isHTTP("https:// this is a random https sentence")).toEqual(false);
      expect(
        isHTTP("ssh://git@github.com/weaveworks/weave-gitops-clusters"),
      ).toEqual(false);
      expect(isHTTP("github.com/weaveworks/weave-gitops-clusters")).toEqual(
        false,
      );
      expect(isHTTP("foo/file.html")).toEqual(false);
      expect(isHTTP("//.com")).toEqual(false);
    });
  });
  describe("isAllowedLink", () => {
    it("allows http", () => {
      expect(isAllowedLink("http://www.google.com")).toEqual(true);
      expect(isAllowedLink("http:// this is a random http sentence")).toEqual(
        true,
      );
    });
    it("allows https", () => {
      expect(isAllowedLink("https://www.google.com")).toEqual(true);
      expect(isAllowedLink("https:// this is a random https sentence")).toEqual(
        true,
      );
    });
    it("allows relative links", () => {
      // Some of these are nonsensical, but if you _want_ them to be a
      // relative link, it's not forbidden.
      expect(isAllowedLink("/hello")).toEqual(true);
      expect(isAllowedLink("test string")).toEqual(true);
      expect(
        isAllowedLink("github.com/weaveworks/weave-gitops-clusters"),
      ).toEqual(true);
      expect(isAllowedLink("foo/file.html")).toEqual(true);
      expect(isAllowedLink("//.com")).toEqual(true);
    });
    it("doesn't allow other links", () => {
      expect(isAllowedLink("oci://server/")).toEqual(false);
      expect(isAllowedLink("smtp://server/")).toEqual(false);
      expect(isAllowedLink("smtp://http/")).toEqual(false);
      expect(isAllowedLink("smtp://https/")).toEqual(false);
      expect(
        isAllowedLink("ssh://git@github.com/weaveworks/weave-gitops-clusters"),
      ).toEqual(false);
    });
  });
  describe("convertGitURLToGitProvider", () => {
    it("converts valid Git URL", () => {
      expect(
        convertGitURLToGitProvider(
          "ssh://git@github.com/weaveworks/weave-gitops-clusters",
        ),
      ).toEqual("https://github.com/weaveworks/weave-gitops-clusters");
    });
    it("returns nothing on invalid Git URL", () => {
      const uri = "github.com/weaveworks/weave-gitops-clusters";

      expect(convertGitURLToGitProvider(uri)).toEqual("");
    });
    it("returns the original HTTP URL", () => {
      expect(
        convertGitURLToGitProvider(
          "https://github.com/weaveworks/weave-gitops-clusters",
        ),
      ).toEqual("https://github.com/weaveworks/weave-gitops-clusters");
    });
  });
  describe("pageTitleWithAppName", () => {
    const pageTitle = "Page Title";
    const appName = "App Name";

    it("returns correct page title with app name", () => {
      expect(pageTitleWithAppName(pageTitle, appName)).toEqual(
        `${pageTitle} for ${appName}`,
      );
    });
    it("returns correct page title without app name", () => {
      expect(pageTitleWithAppName(pageTitle)).toEqual(pageTitle);
    });
  });
  describe("statusSortHelper", () => {
    it("computes suspended status", () => {
      expect(
        statusSortHelper({
          suspended: true,
          conditions: [
            {
              message:
                "Applied revision: main/8868a29b71c008c06549052389f3d762d5fbf821",
              reason: "ReconciliationSucceeded",
              status: "True",
              timestamp: "2022-04-13 20:23:15 +0000 UTC",
              type: "Ready",
            },
          ],
        }),
      ).toEqual(2);
    });
    it("computes ready status", () => {
      expect(
        statusSortHelper({
          suspended: false,
          conditions: [
            {
              type: "Ready",
              status: "True",
              reason: "HealthCheckFailed",
              message:
                "Health check failed after 30.004470633s, timeout waiting for: [Deployment/test/backend status: 'Failed']",
              timestamp: "2022-03-03 16:55:29 +0000 UTC",
            },
            {
              type: "Healthy",
              status: "True",
              reason: "HealthCheckFailed",
              message: "HealthCheckFailed",
              timestamp: "2022-03-03 16:55:29 +0000 UTC",
            },
          ],
        }),
      ).toEqual(4);
    });
    it("computes reconciling status", () => {
      expect(
        statusSortHelper({
          suspended: false,
          conditions: [
            {
              type: "Ready",
              status: "Unknown",
              reason: "Progressing",
              message: "HealthCheckFailed",
              timestamp: "2022-03-03 16:55:29 +0000 UTC",
            },
          ],
        }),
      ).toEqual(3);
    });
    it("computes default status", () => {
      expect(
        statusSortHelper({
          suspended: false,
          conditions: [
            {
              type: "Healthy",
              status: "False",
              reason: "HealthCheckFailed",
              message: "HealthCheckFailed",
              timestamp: "2022-03-03 16:55:29 +0000 UTC",
            },
          ],
        }),
      ).toEqual(1);
    });
  });
  describe("makeImageString", () => {
    it("returns a hyphen if the first image string is empty", () => {
      expect(makeImageString([""])).toEqual("-");
    });
    it("returns the first string if the first string is the only string available and it is not empty", () => {
      expect(makeImageString(["image string 1"])).toEqual("image string 1");
    });
    it("returns the first string if the first string is not empty and the second string is empty", () => {
      expect(makeImageString(["image string 1", ""])).toEqual("image string 1");
    });
    it("concatenates strings if both strings are not empty", () => {
      expect(makeImageString(["image string 1", "image string 2"])).toEqual(
        "image string 1\nimage string 2",
      );
    });
  });
  describe("formatMetadataKey", () => {
    it("capitalizes words in keys if needed", () => {
      expect(formatMetadataKey("description")).toEqual("Description");
      expect(formatMetadataKey("created-by")).toEqual("Created By");
      expect(formatMetadataKey("createdBy")).toEqual("CreatedBy");
      expect(formatMetadataKey("CreatedBy")).toEqual("CreatedBy");
    });
  });
  describe("convertImage", () => {
    it("should handle Docker namespaced repositories", () => {
      expect(convertImage("weaveworks/eksctl")).toEqual(
        "https://hub.docker.com/r/weaveworks/eksctl",
      );
      expect(convertImage("docker.io/weaveworks/eksctl")).toEqual(
        "https://hub.docker.com/r/weaveworks/eksctl",
      );
    });
    it("should handle Docker global repositories", () => {
      expect(convertImage("nginx")).toEqual("https://hub.docker.com/r/_/nginx");
      expect(convertImage("docker.io/nginx")).toEqual(
        "https://hub.docker.com/r/_/nginx",
      );
    });
    it("should handle Docker library alias", () => {
      expect(convertImage("library/nginx")).toEqual(
        "https://hub.docker.com/r/_/nginx",
      );
      expect(convertImage("docker.io/library/nginx")).toEqual(
        "https://hub.docker.com/r/_/nginx",
      );
    });
    it("should handle Quay.io repositories", () => {
      expect(convertImage("quay.io/jitesoft/nginx")).toEqual(
        "https://quay.io/repository/jitesoft/nginx",
      );
    });
    it("should handle Github and Google GHCR/GCR", () => {
      expect(convertImage("ghcr.io/weaveworks/charts/weave-gitops")).toEqual(
        "https://ghcr.io/weaveworks/charts/weave-gitops",
      );
      expect(convertImage("gcr.io/cloud-builders/gcloud")).toEqual(
        "https://gcr.io/cloud-builders/gcloud",
      );
    });
    it("should remove tags", () => {
      expect(
        convertImage("ghcr.io/weaveworks/charts/weave-gitops:10.4.5.2335224"),
      ).toEqual("https://ghcr.io/weaveworks/charts/weave-gitops");
    });
    it("should not link to unsupported images", () => {
      expect(
        convertImage(
          "fakeimage.itisfake.donotdoit.io/fake/fake/fake.com.net.org",
        ),
      ).toEqual(false);
    });
  });
  describe("getSourceRefForAutomation", () => {
    it("should return sourceRef for kustomization", () => {
      const response = {
        payload:
          '{"apiVersion":"kustomize.toolkit.fluxcd.io/v1","kind":"Kustomization","metadata":{"creationTimestamp":"2022-09-14T16:49:20Z","finalizers":["finalizers.fluxcd.io"],"generation":1,"labels":{"kustomize.toolkit.fluxcd.io/name":"kustomization-testdata","kustomize.toolkit.fluxcd.io/namespace":"flux-system"},"name":"backend","namespace":"default","resourceVersion":"293089","uid":"907093ec-5471-46f9-9953-b5b36f9f7859"},"spec":{"dependsOn":[{"name":"common"}],"force":false,"healthChecks":[{"kind":"Deployment","name":"backend","namespace":"webapp"}],"interval":"5m","path":"./deploy/webapp/backend/","prune":true,"sourceRef":{"kind":"GitRepository","name":"webapp"},"timeout":"2m","validation":"server"},"status":{"conditions":[{"lastTransitionTime":"2022-09-15T12:40:22Z","message":"Applied revision: 6.2.0/79f81383288bf6542fcb5bdd8144b826b33b36e7","reason":"ReconciliationSucceeded","status":"True","type":"Ready"},{"lastTransitionTime":"2022-09-15T12:40:22Z","message":"ReconciliationSucceeded","reason":"ReconciliationSucceeded","status":"True","type":"Healthy"}],"inventory":{"entries":[{"id":"webapp_backend__Service","v":"v1"},{"id":"webapp_backend_apps_Deployment","v":"v1"},{"id":"webapp_backend_autoscaling_HorizontalPodAutoscaler","v":"v2beta2"}]},"lastAppliedRevision":"6.2.0/79f81383288bf6542fcb5bdd8144b826b33b36e7","lastAttemptedRevision":"6.2.0/79f81383288bf6542fcb5bdd8144b826b33b36e7","observedGeneration":1}}\n',
        clusterName: "Default",
        tenant: "",
        uid: "907093ec-5471-46f9-9953-b5b36f9f7859",
        inventory: [],
      };

      const kustomization = new Kustomization(response);

      expect(getSourceRefForAutomation(kustomization)).toEqual(
        kustomization.sourceRef,
      );
    });
    it("should return sourceRef for helmrelease", () => {
      const object = {
        payload:
          '{"apiVersion":"helm.toolkit.fluxcd.io/v2","kind":"HelmRelease","metadata":{"annotations":{"reconcile.fluxcd.io/requestedAt":"2022-09-14T14:16:56.304148696Z"},"creationTimestamp":"2022-09-14T14:14:46Z","finalizers":["finalizers.fluxcd.io"],"generation":3,"managedFields":[{"apiVersion":"helm.toolkit.fluxcd.io/v2","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{".":{},"f:chart":{".":{},"f:spec":{".":{},"f:chart":{},"f:reconcileStrategy":{},"f:sourceRef":{".":{},"f:kind":{},"f:name":{}},"f:version":{}}},"f:interval":{},"f:targetNamespace":{}}},"manager":"flux","operation":"Update","time":"2022-09-14T14:14:46Z"},{"apiVersion":"helm.toolkit.fluxcd.io/v2","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:finalizers":{".":{},"v:\\"finalizers.fluxcd.io\\"":{}}}},"manager":"helm-controller","operation":"Update","time":"2022-09-14T14:14:46Z"},{"apiVersion":"helm.toolkit.fluxcd.io/v2","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:reconcile.fluxcd.io/requestedAt":{}}}},"manager":"gitops-server","operation":"Update","time":"2022-09-14T14:17:13Z"},{"apiVersion":"helm.toolkit.fluxcd.io/v2","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:conditions":{},"f:helmChart":{},"f:lastAttemptedRevision":{},"f:lastAttemptedValuesChecksum":{},"f:lastHandledReconcileAt":{},"f:lastReleaseRevision":{},"f:observedGeneration":{}}},"manager":"helm-controller","operation":"Update","subresource":"status","time":"2022-09-14T14:17:20Z"}],"name":"ww-gitops","namespace":"flux-system","resourceVersion":"17512","uid":"2dd24865-4ae4-4a0e-9c78-3204a470be9f"},"spec":{"chart":{"spec":{"chart":"weave-gitops","reconcileStrategy":"ChartVersion","sourceRef":{"kind":"HelmRepository","name":"ww-gitops"},"version":"*"}},"interval":"1m0s","targetNamespace":"weave-gitops"},"status":{"conditions":[{"lastTransitionTime":"2022-09-14T14:17:20Z","message":"Reconciliation in progress","reason":"Progressing","status":"Unknown","type":"Ready"}],"helmChart":"flux-system/flux-system-ww-gitops","lastAttemptedRevision":"4.0.0","lastAttemptedValuesChecksum":"da39a3ee5e6b4b0d3255bfef95601890afd80709","lastHandledReconcileAt":"2022-09-14T14:16:56.304148696Z","lastReleaseRevision":1,"observedGeneration":3}}\n',
        clusterName: "Default",
        tenant: "",
        uid: "2dd24865-4ae4-4a0e-9c78-3204a470be9f",
      };

      const helmRelease = new HelmRelease(object);

      expect(getSourceRefForAutomation(helmRelease)).toEqual(
        helmRelease.helmChart.sourceRef,
      );
    });
    it("should return undefined if automation is undefined", () => {
      let automation: Automation;

      expect(getSourceRefForAutomation(automation)).toBeUndefined();
    });
  });
  describe("getAppVersion", () => {
    const fullResponse: GetVersionResponse = {
      semver: "semver",
      commit: "commit",
      branch: "branch",
      buildTime: "buildTime",
      kubeVersion: "kube-version",
    };
    const defaultVersion = "default version";
    const defaultVersionPrefix = "v";

    it("should return default version for full response if loading data", () => {
      const appVersion = getAppVersion(
        fullResponse,
        defaultVersion,
        true,
        defaultVersionPrefix,
      );

      expect(appVersion.versionText).toEqual(`vdefault version`);
      expect(appVersion.versionHref).toEqual(
        "https://github.com/weaveworks/weave-gitops/releases/tag/vdefault version",
      );
    });
    it("should return api version for full response if not loading data", () => {
      const appVersion = getAppVersion(
        fullResponse,
        defaultVersion,
        false,
        defaultVersionPrefix,
      );

      expect(appVersion.versionText).toEqual("branch-commit");
      expect(appVersion.versionHref).toEqual(
        "https://github.com/weaveworks/weave-gitops/commit/commit",
      );
    });
    it("should return default version without prefix for full response if loading data", () => {
      const appVersion = getAppVersion(fullResponse, defaultVersion, true);

      expect(appVersion.versionText).toEqual(`default version`);
      expect(appVersion.versionHref).toEqual(
        "https://github.com/weaveworks/weave-gitops/releases/tag/vdefault version",
      );
    });
    it("should return api version without prefix for full response", () => {
      const appVersion = getAppVersion(fullResponse, defaultVersion, false);

      expect(appVersion.versionText).toEqual("branch-commit");
      expect(appVersion.versionHref).toEqual(
        "https://github.com/weaveworks/weave-gitops/commit/commit",
      );
    });
  });
  describe("formatLogTimestamp", () => {
    it("should format non-empty timestamp", () => {
      expect(formatLogTimestamp("2023-01-31T13:27:56-05:00", "UTC+1")).toEqual(
        "2023-01-31 19:27:56 UTC+1",
      );
      expect(formatLogTimestamp("2023-02-03T13:27:56-01:00", "UTC-10")).toEqual(
        "2023-02-03 04:27:56 UTC-10",
      );
      expect(formatLogTimestamp("2023-02-03T13:27:56-01:00", "UTC")).toEqual(
        "2023-02-03 14:27:56 UTC",
      );
      expect(formatLogTimestamp("2023-02-04T18:36:01+01:00", "UTC+3")).toEqual(
        "2023-02-04 20:36:01 UTC+3",
      );
    });
    it("should return a hyphen for undefined timestamp", () => {
      expect(formatLogTimestamp(undefined)).toEqual("-");
    });
    it("should return a hyphen for empty string", () => {
      expect(formatLogTimestamp("")).toEqual("-");
    });
  });
});

describe("createYamlCommand", () => {
  it("creates kubectl get yaml string for objects with namespaces", () => {
    expect(
      createYamlCommand(Kind.Kustomization, "test", "flux-system"),
    ).toEqual(`kubectl get kustomization test -n flux-system -o yaml`);
  });
  it("creates kubectl get yaml string for objects without namespaces", () => {
    expect(createYamlCommand(Kind.Kustomization, "test", undefined)).toEqual(
      `kubectl get kustomization test -o yaml`,
    );
  });
  it("returns empty string if name or kind are false values", () => {
    expect(createYamlCommand(undefined, undefined, "flux-system")).toEqual("");
  });
  it("uses the path prop if it is defined", () => {
    expect(createYamlCommand(undefined, undefined, undefined, "http")).toEqual(
      "http",
    );
  });

  describe("getBasePath", () => {
    describe("without a base tag set in the dom", () => {
      let dom: JSDOM;
      beforeEach(() => {
        dom = new JSDOM(
          "<!DOCTYPE html><html><head></head><body></body></html>",
          { url: "https://example.org/" },
        );
      });

      it("should return an empty string", () => {
        expect(getBasePath(dom.window.document)).toEqual("");
      });

      afterEach(() => {
        dom.window.close();
      });
    });

    describe("with a base tag set in the dom", () => {
      let dom: JSDOM;
      beforeEach(() => {
        dom = new JSDOM(
          "<!DOCTYPE html><html><head><base href='/base/'></head><body></body></html>",
          { url: "https://example.org/" },
        );
      });

      it("should return the base URL, stripped of trailing slashes", () => {
        expect(getBasePath(dom.window.document)).toEqual("/base");
      });

      afterEach(() => {
        dom.window.close();
      });
    });
  });
});
