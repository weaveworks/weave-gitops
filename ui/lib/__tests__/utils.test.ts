import { jest } from "@jest/globals";
import { Automation, HelmRelease, Kustomization } from "../objects";
import {
  convertGitURLToGitProvider,
  convertImage,
  formatMetadataKey,
  getSourceRefForAutomation,
  gitlabOAuthRedirectURI,
  isAllowedLink,
  isHTTP,
  makeImageString,
  pageTitleWithAppName,
  statusSortHelper,
} from "../utils";

describe("utils lib", () => {
  describe("gitlabOAuthRedirectURI", () => {
    let windowSpy;

    beforeEach(() => {
      windowSpy = jest.spyOn(window, "window", "get");
    });

    afterEach(() => {
      windowSpy.mockRestore();
    });

    it("returns correct URL", () => {
      windowSpy.mockImplementation(() => ({
        location: {
          origin: "https://example.com",
        },
      }));

      expect(gitlabOAuthRedirectURI()).toEqual(
        "https://example.com/oauth/gitlab"
      );
    });
  });
  describe("isHTTP", () => {
    it("detects HTTP", () => {
      expect(isHTTP("http://www.google.com")).toEqual(true);
      expect(isHTTP("http://www.google.com/")).toEqual(true);
      expect(
        isHTTP("http://github.com/weaveworks/weave-gitops-clusters")
      ).toEqual(true);
      expect(
        isHTTP("http://github.com/weaveworks/weave-gitops-clusters/")
      ).toEqual(true);
    });
    it("detects HTTPS", () => {
      expect(isHTTP("https://www.google.com")).toEqual(true);
      expect(isHTTP("https://www.google.com/")).toEqual(true);
      expect(
        isHTTP("https://github.com/weaveworks/weave-gitops-clusters")
      ).toEqual(true);
      expect(
        isHTTP("https://github.com/weaveworks/weave-gitops-clusters/")
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
        isHTTP("ssh://git@github.com/weaveworks/weave-gitops-clusters")
      ).toEqual(false);
      expect(isHTTP("github.com/weaveworks/weave-gitops-clusters")).toEqual(
        false
      );
      expect(isHTTP("foo/file.html")).toEqual(false);
      expect(isHTTP("//.com")).toEqual(false);
    });
  });
  describe("isAllowedLink", () => {
    it("allows http", () => {
      expect(isAllowedLink("http://www.google.com")).toEqual(true);
      expect(isAllowedLink("http:// this is a random http sentence")).toEqual(
        true
      );
    });
    it("allows https", () => {
      expect(isAllowedLink("https://www.google.com")).toEqual(true);
      expect(isAllowedLink("https:// this is a random https sentence")).toEqual(
        true
      );
    });
    it("allows relative links", () => {
      // Some of these are nonsensical, but if you _want_ them to be a
      // relative link, it's not forbidden.
      expect(isAllowedLink("/hello")).toEqual(true);
      expect(isAllowedLink("test string")).toEqual(true);
      expect(
        isAllowedLink("github.com/weaveworks/weave-gitops-clusters")
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
        isAllowedLink("ssh://git@github.com/weaveworks/weave-gitops-clusters")
      ).toEqual(false);
    });
  });
  describe("convertGitURLToGitProvider", () => {
    it("converts valid Git URL", () => {
      expect(
        convertGitURLToGitProvider(
          "ssh://git@github.com/weaveworks/weave-gitops-clusters"
        )
      ).toEqual("https://github.com/weaveworks/weave-gitops-clusters");
    });
    it("returns nothing on invalid Git URL", () => {
      const uri = "github.com/weaveworks/weave-gitops-clusters";

      expect(convertGitURLToGitProvider(uri)).toEqual("");
    });
    it("returns the original HTTP URL", () => {
      expect(
        convertGitURLToGitProvider(
          "https://github.com/weaveworks/weave-gitops-clusters"
        )
      ).toEqual("https://github.com/weaveworks/weave-gitops-clusters");
    });
  });
  describe("pageTitleWithAppName", () => {
    const pageTitle = "Page Title";
    const appName = "App Name";

    it("returns correct page title with app name", () => {
      expect(pageTitleWithAppName(pageTitle, appName)).toEqual(
        `${pageTitle} for ${appName}`
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
        })
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
        })
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
        })
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
        })
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
        "image string 1\nimage string 2"
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
        "https://hub.docker.com/r/weaveworks/eksctl"
      );
      expect(convertImage("docker.io/weaveworks/eksctl")).toEqual(
        "https://hub.docker.com/r/weaveworks/eksctl"
      );
    });
    it("should handle Docker global repositories", () => {
      expect(convertImage("nginx")).toEqual("https://hub.docker.com/r/_/nginx");
      expect(convertImage("docker.io/nginx")).toEqual(
        "https://hub.docker.com/r/_/nginx"
      );
    });
    it("should handle Docker library alias", () => {
      expect(convertImage("library/nginx")).toEqual(
        "https://hub.docker.com/r/_/nginx"
      );
      expect(convertImage("docker.io/library/nginx")).toEqual(
        "https://hub.docker.com/r/_/nginx"
      );
    });
    it("should handle Quay.io repositories", () => {
      expect(convertImage("quay.io/jitesoft/nginx")).toEqual(
        "https://quay.io/repository/jitesoft/nginx"
      );
    });
    it("should handle Github and Google GHCR/GCR", () => {
      expect(convertImage("ghcr.io/weaveworks/charts/weave-gitops")).toEqual(
        "https://ghcr.io/weaveworks/charts/weave-gitops"
      );
      expect(convertImage("gcr.io/cloud-builders/gcloud")).toEqual(
        "https://gcr.io/cloud-builders/gcloud"
      );
    });
    it("should remove tags", () => {
      expect(
        convertImage("ghcr.io/weaveworks/charts/weave-gitops:10.4.5.2335224")
      ).toEqual("https://ghcr.io/weaveworks/charts/weave-gitops");
    });
    it("should not link to unsupported images", () => {
      expect(
        convertImage(
          "fakeimage.itisfake.donotdoit.io/fake/fake/fake.com.net.org"
        )
      ).toEqual(false);
    });
  });
  describe("getSourceRefForAutomation", () => {
    it("should return sourceRef for kustomization", () => {
      const response = {
        payload:
          '{"apiVersion":"kustomize.toolkit.fluxcd.io/v1beta2","kind":"Kustomization","metadata":{"creationTimestamp":"2022-09-14T16:49:20Z","finalizers":["finalizers.fluxcd.io"],"generation":1,"labels":{"kustomize.toolkit.fluxcd.io/name":"kustomization-testdata","kustomize.toolkit.fluxcd.io/namespace":"flux-system"},"name":"backend","namespace":"default","resourceVersion":"293089","uid":"907093ec-5471-46f9-9953-b5b36f9f7859"},"spec":{"dependsOn":[{"name":"common"}],"force":false,"healthChecks":[{"kind":"Deployment","name":"backend","namespace":"webapp"}],"interval":"5m","path":"./deploy/webapp/backend/","prune":true,"sourceRef":{"kind":"GitRepository","name":"webapp"},"timeout":"2m","validation":"server"},"status":{"conditions":[{"lastTransitionTime":"2022-09-15T12:40:22Z","message":"Applied revision: 6.2.0/79f81383288bf6542fcb5bdd8144b826b33b36e7","reason":"ReconciliationSucceeded","status":"True","type":"Ready"},{"lastTransitionTime":"2022-09-15T12:40:22Z","message":"ReconciliationSucceeded","reason":"ReconciliationSucceeded","status":"True","type":"Healthy"}],"inventory":{"entries":[{"id":"webapp_backend__Service","v":"v1"},{"id":"webapp_backend_apps_Deployment","v":"v1"},{"id":"webapp_backend_autoscaling_HorizontalPodAutoscaler","v":"v2beta2"}]},"lastAppliedRevision":"6.2.0/79f81383288bf6542fcb5bdd8144b826b33b36e7","lastAttemptedRevision":"6.2.0/79f81383288bf6542fcb5bdd8144b826b33b36e7","observedGeneration":1}}\n',
        clusterName: "Default",
        tenant: "",
        uid: "907093ec-5471-46f9-9953-b5b36f9f7859",
        inventory: [],
      };

      const kustomization = new Kustomization(response);

      expect(getSourceRefForAutomation(kustomization)).toBe(
        kustomization.sourceRef
      );
    });
    it("should return sourceRef for helmrelease", () => {
      const object = {
        payload:
          '{"apiVersion":"helm.toolkit.fluxcd.io/v2beta1","kind":"HelmRelease","metadata":{"annotations":{"reconcile.fluxcd.io/requestedAt":"2022-09-14T14:16:56.304148696Z"},"creationTimestamp":"2022-09-14T14:14:46Z","finalizers":["finalizers.fluxcd.io"],"generation":3,"managedFields":[{"apiVersion":"helm.toolkit.fluxcd.io/v2beta1","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{".":{},"f:chart":{".":{},"f:spec":{".":{},"f:chart":{},"f:reconcileStrategy":{},"f:sourceRef":{".":{},"f:kind":{},"f:name":{}},"f:version":{}}},"f:interval":{},"f:targetNamespace":{}}},"manager":"flux","operation":"Update","time":"2022-09-14T14:14:46Z"},{"apiVersion":"helm.toolkit.fluxcd.io/v2beta1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:finalizers":{".":{},"v:\\"finalizers.fluxcd.io\\"":{}}}},"manager":"helm-controller","operation":"Update","time":"2022-09-14T14:14:46Z"},{"apiVersion":"helm.toolkit.fluxcd.io/v2beta1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:reconcile.fluxcd.io/requestedAt":{}}}},"manager":"gitops-server","operation":"Update","time":"2022-09-14T14:17:13Z"},{"apiVersion":"helm.toolkit.fluxcd.io/v2beta1","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:conditions":{},"f:helmChart":{},"f:lastAttemptedRevision":{},"f:lastAttemptedValuesChecksum":{},"f:lastHandledReconcileAt":{},"f:lastReleaseRevision":{},"f:observedGeneration":{}}},"manager":"helm-controller","operation":"Update","subresource":"status","time":"2022-09-14T14:17:20Z"}],"name":"ww-gitops","namespace":"flux-system","resourceVersion":"17512","uid":"2dd24865-4ae4-4a0e-9c78-3204a470be9f"},"spec":{"chart":{"spec":{"chart":"weave-gitops","reconcileStrategy":"ChartVersion","sourceRef":{"kind":"HelmRepository","name":"ww-gitops"},"version":"*"}},"interval":"1m0s","targetNamespace":"weave-gitops"},"status":{"conditions":[{"lastTransitionTime":"2022-09-14T14:17:20Z","message":"Reconciliation in progress","reason":"Progressing","status":"Unknown","type":"Ready"}],"helmChart":"flux-system/flux-system-ww-gitops","lastAttemptedRevision":"4.0.0","lastAttemptedValuesChecksum":"da39a3ee5e6b4b0d3255bfef95601890afd80709","lastHandledReconcileAt":"2022-09-14T14:16:56.304148696Z","lastReleaseRevision":1,"observedGeneration":3}}\n',
        clusterName: "Default",
        tenant: "",
        uid: "2dd24865-4ae4-4a0e-9c78-3204a470be9f",
      };

      const helmRelease = new HelmRelease(object);

      expect(getSourceRefForAutomation(helmRelease)).toBe(
        helmRelease.helmChart.sourceRef
      );
    });
    it("should return undefined if automation is undefined", () => {
      let automation: Automation;

      expect(getSourceRefForAutomation(automation)).toBeUndefined();
    });
  });
});
