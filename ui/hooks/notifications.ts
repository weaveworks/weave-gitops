import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { ListObjectsResponse } from "../lib/api/core/core.pb";
import { Kind } from "../lib/objects";
import { NoNamespace, ReactQueryOptions, RequestError } from "../lib/types";

export function useListProviders(
  namespace: string = NoNamespace,
  opts: ReactQueryOptions<ListObjectsResponse, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);
  return useQuery<ListObjectsResponse, RequestError>(
    ["providers", namespace],
    () => api.ListObjects({ namespace, kind: Kind.Provider }),
    opts
  );
}
