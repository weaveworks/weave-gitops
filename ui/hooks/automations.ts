import _ from "lodash";
import { useContext } from "react";
import { useQuery } from "react-query";
import { AppContext } from "../contexts/AppContext";
import {
  GetHelmReleaseResponse,
  GetKustomizationResponse,
  ListHelmReleasesResponse,
  ListKustomizationsResponse,
} from "../lib/api/core/core.pb";
import { Kustomization } from "../lib/api/core/types.pb";
import { AutomationType, RequestError, WeGONamespace } from "../lib/types";

export type Automation = Kustomization & { type: AutomationType };

export function useListAutomations(namespace = WeGONamespace) {
  const { api } = useContext(AppContext);

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
            type: AutomationType.Kustomization,
          })),
          ..._.map(helmReleases, (h) => ({
            ...h,
            type: AutomationType.HelmRelease,
          })),
        ];
      });
    },
    { retry: false }
  );
}

export function useGetKustomization(name: string, namespace = WeGONamespace) {
  const { api } = useContext(AppContext);

  return useQuery<GetKustomizationResponse, RequestError>(
    ["kustomizations", name],
    () => api.GetKustomization({ name, namespace }),
    { retry: false }
  );
}

export function useGetHelmRelease(name: string, namespace = WeGONamespace) {
  const { api } = useContext(AppContext);

  return useQuery<GetHelmReleaseResponse, RequestError>(
    ["helmrelease", name],
    () => api.GetHelmRelease({ name, namespace }),
    { retry: false }
  );
}
