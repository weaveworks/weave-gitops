import { jest } from "@jest/globals";
import {
  addKind,
  calculateZoomRatio,
  calculateNodeOffsetX,
  convertGitURLToGitProvider,
  convertImage,
  formatMetadataKey,
  gitlabOAuthRedirectURI,
  isHTTP,
  makeImageString,
  mapScaleToZoomPercent,
  mapZoomPercentToScale,
  pageTitleWithAppName,
  removeKind,
  statusSortHelper,
} from "../utils";

const floatPrecision = 19;

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
  describe("convertGitURLToGitProvider", () => {
    it("converts valid Git URL", () => {
      expect(
        convertGitURLToGitProvider(
          "ssh://git@github.com/weaveworks/weave-gitops-clusters"
        )
      ).toEqual("https://github.com/weaveworks/weave-gitops-clusters");
    });
    it("throws error on invalid Git URL", () => {
      const uri = "github.com/weaveworks/weave-gitops-clusters";

      expect(() => {
        convertGitURLToGitProvider(uri);
      }).toThrow(new Error(`could not parse url "${uri}"`));
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
  describe("addKind", () => {
    it("adds the prefix if string does not start with Kind", () => {
      expect(addKind("HelmRelease")).toEqual("KindHelmRelease");
    });
    it("does not add prefix if string starts with Kind", () => {
      expect(addKind("KindGitRepository")).toEqual("KindGitRepository");
    });
  });
  describe("removeKind", () => {
    it("removes the prefix if string starts with Kind", () => {
      expect(removeKind("KindHelmRelease")).toEqual("HelmRelease");
    });
    it("does not remove the prefix if string does not start with Kind", () => {
      expect(removeKind("GitRepository")).toEqual("GitRepository");
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
  describe("calculateZoomRatio", () => {
    it("calculates zoom ratio", () => {
      expect(calculateZoomRatio(0)).toBeCloseTo(
        0.013333333333333334,
        floatPrecision
      );
      expect(calculateZoomRatio(20)).toBeCloseTo(
        0.02666666666666667,
        floatPrecision
      );
      expect(calculateZoomRatio(100)).toBeCloseTo(0.08, floatPrecision);
    });
  });
  describe("calculateNodeOffsetX", () => {
    it("returns 0 if the node is undefined", () => {
      expect(calculateNodeOffsetX(undefined, 20, 0.04)).toEqual(0);
    });

    const rootNode = {
      width: 670,
      x: 3700,
    };
    it("calculates x-offset for the node", () => {
      expect(
        calculateNodeOffsetX(rootNode, 0, 0.013333333333333334)
      ).toBeCloseTo(-40.400000000000006, floatPrecision);
      expect(
        calculateNodeOffsetX(rootNode, 20, 0.02666666666666667)
      ).toBeCloseTo(-55.80000000000001, floatPrecision);
      expect(
        calculateNodeOffsetX(rootNode, 50, 0.04666666666666667)
      ).toBeCloseTo(-78.9, floatPrecision);
      expect(calculateNodeOffsetX(rootNode, 100, 0.08)).toBeCloseTo(
        -117.4,
        floatPrecision
      );
    });
  });
  describe("mapScaleToZoomPercent", () => {
    it("maps zoom percent to scale", () => {
      expect(mapScaleToZoomPercent(0)).toEqual(0);
      expect(mapScaleToZoomPercent(20)).toBeCloseTo(10, floatPrecision);
      expect(mapScaleToZoomPercent(100)).toBeCloseTo(50, floatPrecision);
    });
  });
  describe("mapZoomPercentToScale", () => {
    it("maps scale to zoom percent", () => {
      expect(mapZoomPercentToScale(0)).toEqual(0);
      expect(mapZoomPercentToScale(20)).toBeCloseTo(40, floatPrecision);
      expect(mapZoomPercentToScale(100)).toBeCloseTo(200, floatPrecision);
    });
  });
});
