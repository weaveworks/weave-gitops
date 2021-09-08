// A collection of helper functions to render a graph of kubernetes objects
// with in the context of their parent-child relationships.
import _ from "lodash";
import {
  Application,
  Applications,
  GroupVersionKind,
  UnstructuredObject,
} from "./api/applications/applications.pb";

export type UnstructuredObjectWithParent = UnstructuredObject & {
  parentUid?: string;
};

// Kubernetes does not allow us to query children by parents.
// We keep a list of common parent-child relationships
// to look up children recursively.
export const PARENT_CHILD_LOOKUP = {
  Deployment: {
    group: "apps",
    version: "v1",
    kind: "Deployment",
    children: [
      {
        group: "apps",
        version: "v1",
        kind: "ReplicaSet",
        children: [{ version: "v1", kind: "Pod" }],
      },
    ],
  },
};

export const getChildrenRecursive = async (
  appsClient: typeof Applications,
  result: UnstructuredObjectWithParent[],
  object: UnstructuredObjectWithParent,
  lookup: any
) => {
  result.push(object);

  const k = lookup[object.groupVersionKind.kind];

  if (k && k.children) {
    for (let i = 0; i < k.children.length; i++) {
      const child: GroupVersionKind = k.children[i];

      const res = await appsClient.GetChildObjects({
        parentUid: object.uid,
        groupVersionKind: child,
      });

      for (let q = 0; q < res.objects.length; q++) {
        const c = res.objects[q];

        // Dive down one level and update the lookup accordingly.
        await getChildrenRecursive(
          appsClient,
          result,
          { ...c, parentUid: object.uid },
          {
            [child.kind]: child,
          }
        );
      }
    }
  }
};

// Gets the "child" objects that result from an Application
export const getChildren = async (
  appsClient: typeof Applications,
  app: Application,
  kinds: GroupVersionKind[]
): Promise<UnstructuredObject[]> => {
  const { objects } = await appsClient.GetReconciledObjects({
    automationName: app.name,
    automationNamespace: app.namespace,
    kinds,
  });

  const result = [];
  for (let o = 0; o < objects.length; o++) {
    const obj = objects[o];

    await getChildrenRecursive(appsClient, result, obj, PARENT_CHILD_LOOKUP);
  }

  return _.flatten(result);
};
