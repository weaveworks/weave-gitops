import _ from "lodash";
import { useContext } from "react";
import { useQuery } from "react-query";
import { ReadyStatusValue, ReadyType } from "../components/KubeStatusIndicator";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { ListError } from "../lib/api/core/core.pb";
import { Condition, InventoryEntry } from "../lib/api/core/types.pb";
import { FluxObject } from "../lib/objects";
import { NoNamespace, ReactQueryOptions, RequestError } from "../lib/types";

interface GetInventoryResponse {
  objects?: FluxObject[];
  errors?: ListError[];
}

export function useGetInventory(
  kind: string,
  name: string,
  clusterName: string,
  namespace: string = NoNamespace,
  withChildren?: boolean,
  opts: ReactQueryOptions<GetInventoryResponse, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<GetInventoryResponse, RequestError>(
    ["get_inventory", namespace, kind, withChildren],
    () =>
      api
        .GetInventory({ name, namespace, kind, clusterName, withChildren })
        .then((res) => {
          const listObjects = withChildren
            ? convertEntries(res.entries || [])
            : res.entries?.map((obj) => new FluxObject(obj));
          return { objects: listObjects, errors: [] };
        }),
    opts
  );
}
function convertEntries(entries: InventoryEntry[]) {
  return entries.map((obj) => {
    const parsedObj = new FluxObject(obj);
    const children = obj.children.length ? convertEntries(obj.children) : [];
    parsedObj.children = children;
    return parsedObj;
  });
}

export const createCanaryCondition = (objs: FluxObject[]): Condition => {
  //summarize status of all canaries
  const statusTally = _.reduce(
    objs,
    (prev, current) => {
      if (!current || current.type !== "Canary") return prev;
      const condition = current.conditions[0];
      if (!condition) return prev;
      if (
        condition.reason === "Succeeded" ||
        condition.reason === "Initialized"
      )
        prev["True"] += 1;
      else if (condition.reason === "Failed") prev["False"] += 1;
      else prev["Unknown"] += 1;
      return prev;
    },
    { True: 0, False: 0, Unknown: 0 }
  );
  //create conditions object
  let condition: Condition = {
    type: ReadyType.Ready,
    status: ReadyStatusValue.None,
    message: "No canaries",
  };
  if (statusTally["False"])
    condition = {
      type: ReadyType.Ready,
      status: ReadyStatusValue.False,
      message: "One or more canaries failed",
    };
  else if (statusTally["Unknown"])
    condition = {
      type: ReadyType.Ready,
      status: ReadyStatusValue.Unknown,
      reason: "Progressing",
      message: "One or more canaries are in progress",
    };
  else if (statusTally["True"])
    condition = {
      type: ReadyType.Ready,
      status: ReadyStatusValue.True,
      message: "All canaries have succeeded",
    };
  return condition;
};
