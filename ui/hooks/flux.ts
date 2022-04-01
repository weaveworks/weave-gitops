import { useContext } from "react";
import { useQuery } from "react-query";
import { AppContext } from "../contexts/AppContext";
import { ListFluxRuntimeObjectsResponse } from "../lib/api/core/core.pb";
import {
  AutomationKind,
  GroupVersionKind,
  UnstructuredObject,
} from "../lib/api/core/types.pb";
import { getChildren } from "../lib/graph";
import { RequestError, WeGONamespace, DefaultCluster } from "../lib/types";

export function useListFluxRuntimeObjects(clusterName = DefaultCluster, namespace = WeGONamespace) {
  const { api } = useContext(AppContext);

  return useQuery<ListFluxRuntimeObjectsResponse, RequestError>(
    "flux_runtime_objects",
    () => api.ListFluxRuntimeObjects({ namespace, clusterName }),
    { retry: false, refetchInterval: 5000 }
  );
}

export function useGetReconciledObjects(
  name: string,
  namespace: string,
  type: AutomationKind,
  kinds: GroupVersionKind[],
  clusterName = DefaultCluster,
) {
  const { api } = useContext(AppContext);

  return useQuery<UnstructuredObject[], RequestError>(
    ["reconciled_objects", { name, namespace, type, kinds }],
    () => getChildren(api, name, namespace, kinds, clusterName),
    { retry: false, refetchOnWindowFocus: false, refetchInterval: 5000 }
  );
}
