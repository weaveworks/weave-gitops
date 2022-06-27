import { convertImage } from "../ImageLink";

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
});
