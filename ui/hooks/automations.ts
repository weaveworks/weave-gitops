import _ from "lodash";
import { useContext } from "react";
import { useMutation, useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  GetHelmReleaseResponse,
  GetKustomizationResponse,
  ListHelmReleasesResponse,
  ListKustomizationsResponse,
  SyncAutomationRequest,
  SyncAutomationResponse,
} from "../lib/api/core/core.pb";
import {
  FluxObjectKind,
  HelmRelease,
  Kustomization,
} from "../lib/api/core/types.pb";
import { NoNamespace, RequestError, Syncable } from "../lib/types";

export type Automation = Kustomization & HelmRelease & { kind: FluxObjectKind };

export function useListAutomations(namespace = NoNamespace) {
  const { api } = useContext(CoreClientContext);

  return useQuery<Automation[], RequestError>(
    "automations",
    () => {
      const p = [
        api.ListKustomizations({ namespace }),
        api.ListHelmReleases({ namespace }),
      ];

      // The typescript CLI complains about Promise.all,
      // but VSCode does not. Supress the CLI error here.
      // useQuery will still give us the correct types.
      return Promise.all<any>(p).then((result) => {
        const [kustRes, helmRes] = result;

        const kustomizations = (kustRes as ListKustomizationsResponse)
          .kustomizations;
        const helmReleases = (helmRes as ListHelmReleasesResponse).helmReleases;

        return [
          ..._.map(kustomizations, (k) => ({
            ...k,
            kind: FluxObjectKind.KindKustomization,
          })),
          ..._.map(helmReleases, (h) => ({
            ...h,
            kind: FluxObjectKind.KindHelmRelease,
          })),
        ];
      });
    },
    { retry: false, refetchInterval: 5000 }
  );
}

export function useGetKustomization(
  name: string,

  namespace = NoNamespace,
  clusterName = null
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<GetKustomizationResponse, RequestError>(
    ["kustomizations", name],
    () => api.GetKustomization({ name, namespace, clusterName }),
    { retry: false, refetchInterval: 5000 }
  );
}

export function useGetHelmRelease(
  name: string,
  namespace = NoNamespace,
  clusterName = null
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<GetHelmReleaseResponse, RequestError>(
    ["helmrelease", name],
    () => api.GetHelmRelease({ name, namespace, clusterName }),
    { retry: false, refetchInterval: 5000 }
  );
}

export function useSyncAutomation(obj: Syncable) {
  const { api } = useContext(CoreClientContext);
  const mutation = useMutation<
    SyncAutomationResponse,
    RequestError,
    SyncAutomationRequest
  >(({ withSource }) => api.SyncAutomation({ ...obj, withSource }));

  return mutation;
}
