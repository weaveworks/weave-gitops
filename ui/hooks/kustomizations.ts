import { useContext } from "react";
import { useMutation, useQuery } from "react-query";
import { AppContext } from "../contexts/AppContext";
import {
  AddKustomizationReq,
  AddKustomizationRes,
  ListKustomizationsRes,
} from "../lib/api/app/flux.pb";
import { RequestError } from "../lib/types";

export function useCreateKustomization() {
  const { apps } = useContext(AppContext);

  return useMutation<AddKustomizationRes, RequestError, AddKustomizationReq>(
    (body: AddKustomizationReq) => apps.AddKustomization(body)
  );
}

export function useGetKustomizations(appName: string, namespace: string) {
  const { apps } = useContext(AppContext);

  return useQuery<ListKustomizationsRes, RequestError>(
    ["kustomizations", appName],
    () => apps.ListKustomizations({ appName, namespace }),
    { retry: false }
  );
}
