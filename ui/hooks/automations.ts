import _ from "lodash";
import { useContext } from "react";
import { useQuery } from "react-query";
import { AppContext } from "../contexts/AppContext";
import { GetKustomizationResponse } from "../lib/api/core/core.pb";
import { Kustomization } from "../lib/api/core/types.pb";
import { AutomationType, RequestError, WeGONamespace } from "../lib/types";

export type Automation = Kustomization & { type: AutomationType };

export function useListAutomations(namespace = WeGONamespace) {
  const { api } = useContext(AppContext);

  return useQuery<Automation[], RequestError>(
    "automations",
    () => {
      const p = [api.ListKustomizations({ namespace })];

      return Promise.all(p).then((result) => {
        const [kustRes] = result;

        const kustomizations = kustRes.kustomizations;

        return [
          ..._.map(kustomizations, (k) => ({
            ...k,
            type: AutomationType.Kustomization,
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
