import { findNode, getNeighborNodes, getGraphNodes } from "../dependencies";
import {
  FluxObjectNode,
  Kustomization as FluxObjectKustomization,
} from "../objects";

describe("dependencies", () => {
  const sharedFields = {
    obj: {},
    uid: "some-uid",
    displayKind: "Kustomization",
    suspended: false,
    conditions: [
      {
        type: "Ready",
        status: "True",
        reason: "ReconciliationSucceeded",
        message:
          "Applied revision: main/9e0930cfa1aafef1d8925d2c7b71272b0878aac4",
        timestamp: "2022-09-12T00:31:32Z",
      },
      {
        message: "ReconciliationSucceeded",
        reason: "ReconciliationSucceeded",
        status: "True",
        type: "Healthy",
      },
    ],
    isCurrentNode: false,
  };

  const nodes: FluxObjectNode[] = [
    {
      ...sharedFields,
      name: "kustomization1",
      namespace: "default",
      dependsOn: [],
      parentIds: [],
      isCurrentNode: false,
      id: "default/kustomization1",
    },
    {
      ...sharedFields,
      name: "kustomizationa",
      namespace: "default",
      dependsOn: [],
      parentIds: [],
      isCurrentNode: false,
      id: "default/kustomizationa",
    },
    {
      ...sharedFields,
      name: "kustomizationb",
      namespace: "default",
      dependsOn: [
        {
          name: "kustomizationa",
        },
      ],
      parentIds: ["default/kustomizationa"],
      isCurrentNode: false,
      id: "default/kustomizationb",
    },
    {
      ...sharedFields,
      name: "kustomizationc",
      namespace: "default",
      dependsOn: [
        {
          name: "kustomizationa",
        },
      ],
      parentIds: ["default/kustomizationa"],
      isCurrentNode: false,
      id: "default/kustomizationc",
    },
    {
      ...sharedFields,
      name: "kustomizationd",
      namespace: "default",
      dependsOn: [
        {
          name: "kustomizationb",
        },
      ],
      parentIds: ["default/kustomizationb"],
      isCurrentNode: false,
      id: "default/kustomizationd",
    },
    {
      ...sharedFields,
      name: "kustomizatione",
      namespace: "default",
      dependsOn: [
        {
          name: "kustomizationa",
        },
        {
          name: "kustomization1",
        },
      ],
      parentIds: ["default/kustomizationa", "default/kustomization1"],
      isCurrentNode: false,
      id: "default/kustomizatione",
    },
  ];

  describe("findNode", () => {
    it("returns correct node", () => {
      const node = findNode(nodes, "kustomizationa", "default");

      expect(node).toBe(nodes[1]);
    });
  });
  describe("getNeighborNodes", () => {
    it("returns correct neighbor nodes", () => {
      const node = findNode(nodes, "kustomizationb", "default");

      expect(node).toBe(nodes[2]);

      const neighborNodes = getNeighborNodes(nodes, node);

      expect(neighborNodes.length).toEqual(2);
      expect(neighborNodes[0]).toBe(nodes[1]);
      expect(neighborNodes[1]).toBe(nodes[4]);
    });
  });
  describe("getGraphNodes", () => {
    it("returns correct graph nodes", () => {
      const response = {
        object: {
          payload:
            '{"apiVersion":"kustomize.toolkit.fluxcd.io/v1beta2","kind":"Kustomization","metadata":{"creationTimestamp":"2022-09-11T21:00:14Z","finalizers":["finalizers.fluxcd.io"],"generation":1,"labels":{"kustomize.toolkit.fluxcd.io/name":"kustomization-testdependencies","kustomize.toolkit.fluxcd.io/namespace":"flux-system"},"name":"kustomizationc","namespace":"default","resourceVersion":"157207","uid":"5ae733f8-c4dd-4846-a4dd-c4d5c4b01340"},"spec":{"dependsOn":[{"name":"kustomizationa"}],"force":false,"healthChecks":[{"kind":"Deployment","name":"frontend","namespace":"webapp"}],"interval":"5m","path":"./deploy/webapp/frontend/","prune":true,"sourceRef":{"kind":"GitRepository","name":"webapp"},"timeout":"2m","validation":"server"},"status":{"conditions":[{"lastTransitionTime":"2022-09-12T02:01:00Z","message":"Applied revision: 6.2.0/79f81383288bf6542fcb5bdd8144b826b33b36e7","reason":"ReconciliationSucceeded","status":"True","type":"Ready"},{"lastTransitionTime":"2022-09-12T02:01:00Z","message":"ReconciliationSucceeded","reason":"ReconciliationSucceeded","status":"True","type":"Healthy"}],"inventory":{"entries":[{"id":"webapp_frontend__Service","v":"v1"},{"id":"webapp_frontend_apps_Deployment","v":"v1"},{"id":"webapp_frontend_autoscaling_HorizontalPodAutoscaler","v":"v2beta2"}]},"lastAppliedRevision":"6.2.0/79f81383288bf6542fcb5bdd8144b826b33b36e7","lastAttemptedRevision":"6.2.0/79f81383288bf6542fcb5bdd8144b826b33b36e7","observedGeneration":1}}',
          clusterName: "Default",
        },
      };

      const kustomization = new FluxObjectKustomization(response.object);
      expect(kustomization.name).toEqual("kustomizationc");
      expect(kustomization.namespace).toEqual("default");

      const graphNodes = getGraphNodes(nodes, kustomization);

      expect(graphNodes.length).toEqual(6);
      expect(graphNodes[0]).toBe(nodes[3]);
      expect(graphNodes[1]).toBe(nodes[1]);
      expect(graphNodes[2]).toBe(nodes[2]);
      expect(graphNodes[3]).toBe(nodes[5]);
      expect(graphNodes[4]).toBe(nodes[4]);
      expect(graphNodes[5]).toBe(nodes[0]);
    });
  });
});
