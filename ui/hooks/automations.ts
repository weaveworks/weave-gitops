import _ from "lodash";
import { useContext } from "react";
import { useMutation, useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  ListHelmReleasesResponse,
  ListKustomizationsResponse,
  SyncFluxObjectRequest,
  SyncFluxObjectResponse,
} from "../lib/api/core/core.pb";
import {
  FluxObjectKind,
  HelmRelease,
  Kustomization,
} from "../lib/api/core/types.pb";
import {
  MultiRequestError,
  NoNamespace,
  ReactQueryOptions,
  RequestError,
  Syncable,
} from "../lib/types";
import { notifyError, notifySuccess } from "../lib/utils";

export type Automation = (Kustomization | HelmRelease) & {
  kind: FluxObjectKind;
};

type Res = { result: Automation[]; errors: MultiRequestError[] };

export function useListAutomations(
  namespace = NoNamespace,
  opts: ReactQueryOptions<Res, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<Res, RequestError>(
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
        return {
          result: [
            ..._.map(kustomizations, (k) => ({
              ...k,
              kind: FluxObjectKind.KindKustomization,
            })),
            ..._.map(helmReleases, (h) => ({
              ...h,
              kind: FluxObjectKind.KindHelmRelease,
            })),
          ],
          errors: [
            ..._.map(kustRes.errors, (e) => ({
              ...e,
              kind: FluxObjectKind.KindKustomization,
            })),
            ..._.map(helmRes.errors, (e) => ({
              ...e,
              kind: FluxObjectKind.KindHelmRelease,
            })),
          ],
        };
      });
    },
    opts
  );
}

export function useSyncFluxObject(objs: Syncable[]) {
  const { api } = useContext(CoreClientContext);
  const mutation = useMutation<
    SyncFluxObjectResponse,
    RequestError,
    SyncFluxObjectRequest
  >(({ withSource }) => api.SyncFluxObject({ objects: objs, withSource }), {
    onSuccess: () => notifySuccess("Sync request successful!"),
    onError: (error) => notifyError(error.message),
  });

  return mutation;
}
