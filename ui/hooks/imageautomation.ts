import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { ListObjectsResponse } from "../lib/api/core/core.pb";
import { Kind } from "../lib/api/core/types.pb";
import { NoNamespace, ReactQueryOptions, RequestError } from "../lib/types";
import { convertResponse } from "./objects";

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
    () =>
      api.ListObjects({ namespace, kind }).then((res) => {
        const providers = res.objects?.map((obj) =>
          convertResponse(Kind.ImageRepository, obj)
        );
        return { objects: providers, errors: res.errors };
      }),
    opts
  );
}
