import "jest-styled-components";
import { makeObjects } from "../CheckboxActions";

describe("CheckboxActions", () => {
  it("creates suspend reqs based on props", () => {
    const checked = ["123"];
    const rows = [
      {
        name: "name",
        kind: "kind",
        namespace: "namespace",
        clusterName: "clusterName",
        uid: "123",
      },
      {
        name: "name",
        kind: "kind",
        namespace: "namespace",
        clusterName: "clusterName",
        uid: "321",
      },
    ];
    expect(makeObjects(checked, rows)).toEqual([
      {
        name: "name",
        kind: "kind",
        namespace: "namespace",
        clusterName: "clusterName",
      },
    ]);
  });
});
