import { useContext } from "react";
import { useMutation, useQuery, useQueryClient } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  ListFluxCrdsResponse,
  ListFluxRuntimeObjectsResponse,
  ToggleSuspendResourceRequest,
  ToggleSuspendResourceResponse,
} from "../lib/api/core/core.pb";
import {
  FluxObjectKind,
  GroupVersionKind,
  UnstructuredObject,
} from "../lib/api/core/types.pb";
import { getChildren, UnstructuredObjectWithChildren } from "../lib/graph";
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
  }
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<ListFluxRuntimeObjectsResponse, RequestError>(
    "flux_runtime_objects",
    () => api.ListFluxRuntimeObjects({ namespace, clusterName }),
    opts
  );
}

export function useListFluxCrds(clusterName = DefaultCluster) {
  const { api } = useContext(CoreClientContext);

  return useQuery<ListFluxCrdsResponse, RequestError>(
    "flux_crds",
    () => api.ListFluxCrds({ clusterName }),
    { retry: false, refetchInterval: 5000 }
  );
}

export function flattenChildren(children: UnstructuredObjectWithChildren[]) {
  return children.flatMap((child) => [child].concat(flattenChildren(child.children)))
}

export function useGetReconciledObjects(
  name: string,
  namespace: string,
  type: FluxObjectKind,
  kinds: GroupVersionKind[],
  clusterName = DefaultCluster,
  opts: ReactQueryOptions<UnstructuredObject[], RequestError> = {
    retry: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<UnstructuredObject[], RequestError>(
    ["reconciled_objects", { name, namespace, type, kinds }],
    async () => {
      const childrenTrees = await getChildren(api, name, namespace, type, kinds, clusterName);
      return flattenChildren(childrenTrees)
    },
    opts
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
        const suspend = req.suspend ? "Suspend" : "Resume";
        notifySuccess(`${suspend} request successful!`);
        return queryClient.invalidateQueries(type);
      },
      onError: (error) => {
        notifyError(error.message);
      },
    }
  );
  return mutation;
}
