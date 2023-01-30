import _ from "lodash";
import { Core, GetChildObjectsResponse } from "../api/core/core.pb";
import { Kind } from "../api/core/types.pb";
import { getChildren } from "../graph";
import { createCoreMockClient } from "../test-utils";

describe("graph lib", () => {
  let client: typeof Core;

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
  const rsPayload = {
    groupVersionKind: {
      group: "apps",
      kind: "ReplicaSet",
      version: "v1",
    },
    name: rsName,
    namespace: "default",
    status: "InProgress",
    uid: "70c0f983-f9a4-4375-adfe-c2c018fc10bd",
  };
  const obj2 = {
    payload: JSON.stringify(rsPayload),
  };
  const podPayload = {
    groupVersionKind: {
      group: "",
      kind: "Pod",
      version: "v1",
    },
    name: podName,
    namespace: "default",
    status: "InProgress",
    uid: "70c0f983-f9a4-4375-adfe-c2c018fc10bd",
  };
  const obj3 = {
    payload: JSON.stringify(podPayload),
  };

  beforeEach(() => {
    client = createCoreMockClient({
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
  });
  it("getChildren", async () => {
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
    const rs = _.find(
      dep.children,
      (o) => o.obj.groupVersionKind.kind === "ReplicaSet"
    );
    expect(rs.obj.name).toEqual(rsName);
    expect(rs.obj.groupVersionKind.kind).toEqual("ReplicaSet");
    const pod = _.find(
      rs.children,
      (o) => o.obj.groupVersionKind.kind === "Pod"
    );
    expect(pod).toBeTruthy();
    expect(pod.obj.name).toEqual(podName);
  });
  describe("deterministic graph", () => {
    // https://github.com/weaveworks/weave-gitops/issues/3302
    // Make sure the graph nodes don't "hop around" when the server returns objects in a new order
    beforeEach(() => {
      client = createCoreMockClient({
        GetReconciledObjects: () => {
          return {
            objects: [obj1],
          };
        },
        GetChildObjects: (req) => {
          if (req.groupVersionKind.kind === "ReplicaSet") {
            return {
              objects: [
                obj2,
                { payload: JSON.stringify({ ...rsPayload, name: "other-rs" }) },
              ],
            };
          }

          if (req.groupVersionKind.kind === "Pod") {
            return {
              objects: [
                obj3,
                {
                  payload: JSON.stringify({ ...podPayload, name: "other-pod" }),
                },
              ],
            };
          }
        },
      });
    });
    it("returns children in the same order every time", async () => {
      // https://github.com/weaveworks/weave-gitops/issues/3302
      const objects = await getChildren(
        client,
        app.name,
        app.namespace,
        app.automationKind,
        [{ group: "apps", version: "v1", kind: "Deployment" }],
        app.clusterName
      );

      const firstPods = _.get(objects[0], ["children", 0, "children"]);
      expect(firstPods.length).toEqual(2);

      // Simulate the server returning children in a different order
      const newClient = {
        ...client,
        GetChildObjects: async (req) => {
          const res = await client.GetChildObjects(req);
          // Changing the order here
          const reversed = res.objects.reverse();
          return new Promise<GetChildObjectsResponse>((accept) =>
            accept({ objects: reversed })
          );
        },
      };

      const objects2 = await getChildren(
        // @ts-ignore
        newClient,
        app.name,
        app.namespace,
        app.automationKind,
        [{ group: "apps", version: "v1", kind: "Deployment" }],
        app.clusterName
      );

      const secondPods = _.get(objects2[0], ["children", 0, "children"]);
      expect(secondPods.length).toEqual(2);

      // This will do a deep equal by value
      expect(firstPods).toEqual(secondPods);
    });
    it("returns the same list at every child level", async () => {
      // Ensure that each level of the graph is deterministic.
      // For example, ReplicaSets need to be ordered, as do their child Pods.
      const objects = await getChildren(
        client,
        app.name,
        app.namespace,
        app.automationKind,
        [{ group: "apps", version: "v1", kind: "Deployment" }],
        app.clusterName
      );

      const firstReplicaSets = _.get(objects[0], ["children"]);
      expect(firstReplicaSets).toHaveLength(2);
      const firstPods = _.get(objects[0], ["children", 0, "children"]);
      expect(firstPods).toHaveLength(2);

      // Simulate the server returning children in a different order
      const newClient = {
        ...client,
        GetChildObjects: async (req) => {
          const res = await client.GetChildObjects(req);
          // Changing the order here

          const reversed = res.objects.reverse();
          return new Promise<GetChildObjectsResponse>((accept) =>
            accept({ objects: reversed })
          );
        },
      };

      const objects2 = await getChildren(
        // @ts-ignore
        newClient,
        app.name,
        app.namespace,
        app.automationKind,
        [{ group: "apps", version: "v1", kind: "Deployment" }],
        app.clusterName
      );

      const secondReplicaSets = _.get(objects2[0], ["children"]);
      expect(secondReplicaSets).toHaveLength(2);
      const secondPods = _.get(objects2[0], ["children", 0, "children"]);
      expect(secondPods).toHaveLength(2);

      expect(firstReplicaSets).toEqual(secondReplicaSets);
      expect(firstPods).toEqual(secondPods);
    });
  });
});
