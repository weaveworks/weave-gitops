import { useContext } from "react";
import { useMutation, useQuery, useQueryClient } from "react-query";
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
    { retry: false }
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
    { retry: false, refetchOnWindowFocus: false }
  );
}

export function useToggleSuspend(
  req: ToggleSuspendResourceRequest,
  type: string
) {
  const { api } = useContext(CoreClientContext);
  const queryClient = useQueryClient();
  const mutation = useMutation<ToggleSuspendResourceResponse, RequestError>(
    () => api.ToggleSuspendResource(req),
    {
      onSuccess: () => {
        return queryClient.invalidateQueries(type);
      },
    }
  );
  return mutation;
}
