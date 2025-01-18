import { useQuery } from "@tanstack/react-query";
import { useContext } from "react";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { ListObjectsResponse } from "../lib/api/core/core.pb";
import { Kind } from "../lib/api/core/types.pb";
import { Source } from "../lib/objects";
import {
  MultiRequestError,
  NoNamespace,
  ReactQueryOptions,
  RequestError,
  SearchedNamespaces,
} from "../lib/types";
import { convertResponse } from "./objects";

type Res = {
  result: Source[];
  errors: MultiRequestError[];
  searchedNamespaces: SearchedNamespaces;
};

export function useListSources(
  appName?: string,
  namespace: string = NoNamespace,
  opts: ReactQueryOptions<Res, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  },
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<Res, RequestError>({
    queryKey: ["sources", namespace],
    queryFn: () => {
      const p = [
        Kind.GitRepository,
        Kind.HelmRepository,
        Kind.Bucket,
        Kind.HelmChart,
        Kind.OCIRepository,
      ].map((kind) =>
        api
          .ListObjects({ namespace, kind })
          .then((response: ListObjectsResponse) => {
            if (!response.objects) response.objects = [];
            if (!response.errors) response.errors = [];
            return { kind, response };
          }),
      );
      return Promise.all(p).then((responses) => {
        const final: Res = { result: [], errors: [], searchedNamespaces: {} };
        for (const { kind, response } of responses) {
          final.result.push(
            ...response.objects.map((o) => convertResponse(kind, o) as Source),
          );
          if (response.errors.length) {
            final.errors.push(
              ...response.errors.map((o) => {
                return { ...o, kind };
              }),
            );
          }
          final.searchedNamespaces[kind] = response.searchedNamespaces;
        }
        return final;
      });
    },
    ...opts,
  });
}
