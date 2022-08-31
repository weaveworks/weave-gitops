import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { ListError } from "../lib/api/core/core.pb";
import { Kind, Provider } from "../lib/objects";
import { NoNamespace, ReactQueryOptions, RequestError } from "../lib/types";
import { convertResponse } from "./objects";

type Res = { objects: Provider[]; errors: ListError[] };

export function useListProviders(
  namespace: string = NoNamespace,
  opts: ReactQueryOptions<Res, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);
  return useQuery<Res, RequestError>(
    ["providers", namespace],
    () => {
      return api.ListObjects({ namespace, kind: Kind.Provider }).then((res) => {
        const providers = res.objects?.map(
          (obj) => convertResponse(Kind.Provider, obj) as Provider
        );
        return { objects: providers, errors: res.errors };
      });
    },
    opts
  );
}
