import { useContext } from "react";
import { useMutation, useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  ListFluxRuntimeObjectsResponse,
  ToggleSuspendResourceRequest,
  ToggleSuspendResourceResponse,
} from "../lib/api/core/core.pb";
import {
  FluxObjectKind,
  GroupVersionKind,
  UnstructuredObject,
} from "../lib/api/core/types.pb";
import { getChildren } from "../lib/graph";
import { DefaultCluster, NoNamespace, RequestError } from "../lib/types";

export function useListFluxRuntimeObjects(
  clusterName = DefaultCluster,
  namespace = NoNamespace
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<ListFluxRuntimeObjectsResponse, RequestError>(
    "flux_runtime_objects",
    () => api.ListFluxRuntimeObjects({ namespace, clusterName }),
    { retry: false, refetchInterval: 5000 }
  );
}

export function useGetReconciledObjects(
  name: string,
  namespace: string,
  type: FluxObjectKind,
  kinds: GroupVersionKind[],
  clusterName = DefaultCluster
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<UnstructuredObject[], RequestError>(
    ["reconciled_objects", { name, namespace, type, kinds }],
    () => getChildren(api, name, namespace, type, kinds, clusterName),
    { retry: false, refetchOnWindowFocus: false, refetchInterval: 5000 }
  );
}

export function useToggleSuspend(req: ToggleSuspendResourceRequest) {
  const { api } = useContext(CoreClientContext);
  const mutation = useMutation<ToggleSuspendResourceResponse, RequestError>(
    () => api.ToggleSuspendResource(req)
  );
  return mutation;
}
