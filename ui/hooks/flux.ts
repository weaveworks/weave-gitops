import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useContext } from "react";
import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  ListFluxCrdsResponse,
  ListFluxRuntimeObjectsResponse,
  ListRuntimeObjectsResponse,
  ToggleSuspendResourceRequest,
  ToggleSuspendResourceResponse,
} from "../lib/api/core/core.pb";
import { GroupVersionKind, Kind } from "../lib/api/core/types.pb";
import { getChildren } from "../lib/graph";
import { FluxObject } from "../lib/objects";
import {
  DefaultCluster,
  NoNamespace,
  ReactQueryOptions,
  RequestError,
} from "../lib/types";
import { notifyError, notifySuccess } from "../lib/utils";
export function useListFluxRuntimeObjects(
  clusterName = DefaultCluster,
  namespace = NoNamespace,
  opts: ReactQueryOptions<ListFluxRuntimeObjectsResponse, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  },
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<ListFluxRuntimeObjectsResponse, RequestError>({
    queryKey: ["flux_runtime_objects"],
    queryFn: () => api.ListFluxRuntimeObjects({ namespace, clusterName }),
    ...opts,
  });
}

export function useListFluxCrds(clusterName = DefaultCluster) {
  const { api } = useContext(CoreClientContext);

  return useQuery<ListFluxCrdsResponse, RequestError>({
    queryKey: ["flux_crds"],
    queryFn: () => api.ListFluxCrds({ clusterName }),
    retry: false,
    refetchInterval: 5000,
  });
}

export function useListRuntimeObjects(
  clusterName = DefaultCluster,
  namespace = NoNamespace,
  opts: ReactQueryOptions<ListRuntimeObjectsResponse, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  },
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<ListRuntimeObjectsResponse, RequestError>({
    queryKey: ["runtime_objects"],
    queryFn: () => api.ListRuntimeObjects({ namespace, clusterName }),
    ...opts,
  });
}

export function useListRuntimeCrds(clusterName = DefaultCluster) {
  const { api } = useContext(CoreClientContext);

  return useQuery<ListFluxCrdsResponse, RequestError>({
    queryKey: ["runtime_crds"],
    queryFn: () => api.ListRuntimeCrds({ clusterName }),
    retry: false,
    refetchInterval: 5000,
  });
}

export function flattenChildren(children: FluxObject[]) {
  return children.flatMap((child) =>
    [child].concat(flattenChildren(child.children)),
  );
}

export function useGetReconciledObjects(
  name: string,
  namespace: string,
  type: Kind,
  kinds: GroupVersionKind[],
  clusterName = DefaultCluster,
  opts: ReactQueryOptions<FluxObject[], RequestError> = {
    retry: false,
    refetchInterval: 5000,
  },
) {
  const result = useGetReconciledTree(
    name,
    namespace,
    type,
    kinds,
    clusterName,
    opts,
  );
  if (result.data) {
    result.data = flattenChildren(result.data);
  }
  return result;
}

export function useGetReconciledTree(
  name: string,
  namespace: string,
  type: Kind,
  kinds: GroupVersionKind[],
  clusterName = DefaultCluster,
  opts: ReactQueryOptions<FluxObject[], RequestError> = {
    retry: false,
    refetchInterval: 5000,
  },
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<FluxObject[], RequestError>({
    queryKey: ["reconciled_objects", { name, namespace, type, kinds }],
    queryFn: () => getChildren(api, name, namespace, type, kinds, clusterName),
    ...opts,
  });
}

export function useToggleSuspend(
  req: ToggleSuspendResourceRequest,
  type: string,
) {
  const { api } = useContext(CoreClientContext);
  const queryClient = useQueryClient();
  const mutation = useMutation<ToggleSuspendResourceResponse, RequestError>({
    mutationFn: () => api.ToggleSuspendResource(req),
    onSuccess: () => {
      const suspend = req.suspend ? "Suspend" : "Resume";
      notifySuccess(`${suspend} request successful!`);
      return queryClient.invalidateQueries({ queryKey: [type] });
    },
    onError: (error) => {
      notifyError(error.message);
    },
  });
  return mutation;
}
