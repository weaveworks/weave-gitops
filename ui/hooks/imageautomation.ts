import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  GetInventoryResponse,
  ListObjectsResponse,
} from "../lib/api/core/core.pb";
import { InventoryEntry } from "../lib/api/core/types.pb";
import { FluxObject } from "../lib/objects";
import { NoNamespace, ReactQueryOptions, RequestError } from "../lib/types";
import { convertResponse } from "./objects";

export function useListImageAutomation(
  kind: string,
  namespace: string = NoNamespace,
  opts: ReactQueryOptions<ListObjectsResponse, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<ListObjectsResponse, RequestError>(
    ["image_automation", namespace, kind],
    () =>
      api.ListObjects({ namespace, kind }).then((res) => {
        const providers = res.objects?.map((obj) => convertResponse(kind, obj));
        return { objects: providers, errors: res.errors };
      }),
    opts
  );
}

export function useCheckCRDInstalled(
  name: string,
  opts: ReactQueryOptions<boolean, RequestError> = {
    retry: false,
    refetchInterval: (data) => (data ? false : 5000),
  }
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<boolean, RequestError>(
    ["image_automation_crd_available", name],
    () =>
      api.IsCRDAvailable({ name }).then(({ clusters }) => {
        return Object.values(clusters).some((r) => r === true);
      }),
    opts
  );
}

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
    const { name, namespace, suspended, clusterName, type, uid, tenant } =
      parsedObj;

    return {
      ...parsedObj,
      children: children,
      conditions: parsedObj.conditions,
      name,
      namespace,
      suspended,
      clusterName,
      type,
      uid,
      tenant,
    };
  });
}
