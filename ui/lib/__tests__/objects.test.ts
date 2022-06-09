import { FluxObject } from "../objects";

describe("objects lib", () => {
  it("extracts annotations", () => {
    const payload =
      '{"apiVersion":"helm.toolkit.fluxcd.io/v2beta1","kind":"HelmRelease","metadata":{"annotations":{"metadata.weave.works/test":"value","reconcile.fluxcd.io/requestedAt":"2022-05-25T10:54:36.83322005Z"},"creationTimestamp":"2022-05-24T18:14:46Z","finalizers":["finalizers.fluxcd.io"],"generation":5,"name":"normal","namespace":"a-namespace","resourceVersion":"3978798","uid":"82231842-2224-4f22-8576-5babf08d746d"}}\n';

    const obj = new FluxObject({
      payload,
    });

    const metadata = obj.metadata();
    expect(metadata).toEqual([["test", "value"]]);
  });

  it("doesn't format annotations", () => {
    const payload =
      '{"metadata":{"annotations":{"metadata.weave.works/impolite-value":"<script>alert()</script>\\n"}}}\n';

    const obj = new FluxObject({
      payload,
    });

    const metadata = obj.metadata();
    expect(metadata).toEqual([
      ["impolite-value", "<script>alert()</script>\n"],
    ]);
  });

  it("dumps yaml", () => {
    const payload =
      '{"apiVersion":"helm.toolkit.fluxcd.io/v2beta1","kind":"HelmRelease"}\n';

    const obj = new FluxObject({
      payload,
    });

    const yaml = obj.yaml();
    expect(yaml).toEqual(
      "apiVersion: helm.toolkit.fluxcd.io/v2beta1\nkind: HelmRelease\n"
    );
  });
});
