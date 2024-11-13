import "jest-styled-components";
import { Kind } from "../../../lib/api/core/types.pb";
import { makeObjects } from "../CheckboxActions";

describe("CheckboxActions", () => {
  it("creates suspend reqs based on props", () => {
    const checked = ["123"];
    const rows = [
      {
        name: "name",
        type: Kind.GitRepository,
        namespace: "namespace",
        clusterName: "clusterName",
        uid: "123",
      },
      {
        name: "name",
        type: Kind.HelmRelease,
        namespace: "namespace",
        clusterName: "clusterName",
        uid: "321",
      },
    ];
    expect(makeObjects(checked, rows)).toEqual([
      {
        name: "name",
        kind: Kind.GitRepository,
        namespace: "namespace",
        clusterName: "clusterName",
      },
    ]);
  });
});
