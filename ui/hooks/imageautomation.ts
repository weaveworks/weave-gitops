import { useQuery } from "@tanstack/react-query";
import { useContext } from "react";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { ListObjectsResponse } from "../lib/api/core/core.pb";
import { NoNamespace, ReactQueryOptions, RequestError } from "../lib/types";
import { convertResponse } from "./objects";

export function useListImageAutomation(
  kind: string,
  namespace: string = NoNamespace,
  opts: ReactQueryOptions<ListObjectsResponse, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  },
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<ListObjectsResponse, RequestError>({
    queryKey: ["image_automation", namespace, kind],
    queryFn: () =>
      api.ListObjects({ namespace, kind }).then((res) => {
        const providers = res.objects?.map((obj) => convertResponse(kind, obj));
        return { objects: providers, errors: res.errors };
      }),
    ...opts,
  });
}

export function useCheckCRDInstalled(
  name: string,
  opts: ReactQueryOptions<boolean, RequestError> = {
    retry: false,
    refetchInterval: (data) => (data ? false : 5000),
  },
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<boolean, RequestError>({
    queryKey: ["image_automation_crd_available", name],
    queryFn: () =>
      api.IsCRDAvailable({ name }).then(({ clusters }) => {
        return Object.values(clusters).some((r) => r === true);
      }),
    ...opts,
  });
}
