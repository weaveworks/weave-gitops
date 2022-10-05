import { Kind } from "../api/core/types.pb";
import { getChildren } from "../graph";
import { createCoreMockClient } from "../test-utils";

describe("graph lib", () => {
  it("getChildren", async () => {
    const app = {
      name: "my-app",
      namespace: "my-namespace",
      automationKind: Kind.HelmRelease,
      reconciledObjectKinds: [
        { group: "apps", version: "v1", kind: "Deployment" },
      ],
      clusterName: "foo",
    };
    const name = "stringly";
    const rsName = name + "-7d9b7454c7";
    const podName = rsName + "-mvz75";
    const obj1 = {
      payload: JSON.stringify({
        groupVersionKind: {
          group: "apps",
          kind: "Deployment",
          version: "v1",
        },
        name,
        namespace: "default",
        status: "Failed",
        uid: "2f5b0538-919d-4700-8f41-31eb5e1d9a78",
      }),
    };
    const obj2 = {
      payload: JSON.stringify({
        groupVersionKind: {
          group: "apps",
          kind: "ReplicaSet",
          version: "v1",
        },
        name: rsName,
        namespace: "default",
        status: "InProgress",
        uid: "70c0f983-f9a4-4375-adfe-c2c018fc10bd",
      }),
    };
    const obj3 = {
      payload: JSON.stringify({
        groupVersionKind: {
          group: "",
          kind: "Pod",
          version: "v1",
        },
        name: podName,
        namespace: "default",
        status: "InProgress",
        uid: "70c0f983-f9a4-4375-adfe-c2c018fc10bd",
      }),
    };
    const client = createCoreMockClient({
      GetReconciledObjects: () => {
        return {
          objects: [obj1],
        };
      },
      GetChildObjects: (req) => {
        if (req.groupVersionKind.kind === "ReplicaSet") {
          return {
            objects: [obj2],
          };
        }

        if (req.groupVersionKind.kind === "Pod") {
          return {
            objects: [obj3],
          };
        }
      },
    });

    const objects = await getChildren(
      client,
      app.name,
      app.namespace,
      app.automationKind,
      [{ group: "apps", version: "v1", kind: "Deployment" }],
      app.clusterName
    );
    const dep = objects[0];
    expect(dep).toBeTruthy();
    expect(dep.obj.name).toEqual(name);
    const rs = objects[0].children[0];
    expect(rs.obj.name).toEqual(rsName);
    expect(rs.obj.groupVersionKind.kind).toEqual("ReplicaSet");
    const pod = objects[0].children[0].children[0];
    expect(pod).toBeTruthy();
    expect(pod.obj.name).toEqual(podName);
  });
});
