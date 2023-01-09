import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { ListObjectsResponse } from "../lib/api/core/core.pb";
import {
    NoNamespace,
    ReactQueryOptions,
    RequestError
} from "../lib/types";


export function useListImageAutomation(
  kind: string,
  namespace: string = NoNamespace,
  opts: ReactQueryOptions<ListObjectsResponse, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<ListObjectsResponse, RequestError>(
    ["image_automation", namespace],
    () => api.ListObjects({ namespace, kind }),
    opts
  );
}
