import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { ListObjectsResponse } from "../lib/api/core/core.pb";
import { InventoryEntry } from "../lib/api/core/types.pb";
import { FluxObject } from "../lib/objects";
import { NoNamespace, ReactQueryOptions, RequestError } from "../lib/types";

export function useGetInventory(
  kind: string,
  name: string,
  clusterName: string,
  namespace: string = NoNamespace,
  withChildren?: boolean,
  opts: ReactQueryOptions<ListObjectsResponse, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<ListObjectsResponse, RequestError>(
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
