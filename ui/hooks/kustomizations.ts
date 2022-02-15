import _ from "lodash";
import { useContext } from "react";
import { useMutation, useQuery } from "react-query";
import { AppContext } from "../contexts/AppContext";
import {
  AddKustomizationReq,
  AddKustomizationRes,
  Kustomization,
  ListKustomizationsRes,
} from "../lib/api/app/flux.pb";
import { AutomationType, RequestError, WeGONamespace } from "../lib/types";

export function useCreateKustomization() {
  const { apps } = useContext(AppContext);

  return useMutation<AddKustomizationRes, RequestError, AddKustomizationReq>(
    (body: AddKustomizationReq) => apps.AddKustomization(body)
  );
}

export function useGetKustomizations(
  appName?: string,
  namespace: string = WeGONamespace
) {
  const { apps } = useContext(AppContext);

  return useQuery<ListKustomizationsRes, RequestError>(
    ["kustomizations", appName],
    () => apps.ListKustomizations({ appName, namespace }),
    { retry: false }
  );
}
export type Automation = Kustomization & { type: AutomationType };

export function useListAutomations(namespace = WeGONamespace) {
  const { apps } = useContext(AppContext);

  return useQuery<Automation[], RequestError>(
    "automations",
    () => {
      const p = [apps.ListKustomizations({ namespace })];

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
