import { useContext } from "react";
import { useMutation, useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  ListObjectsResponse,
  SyncFluxObjectRequest,
  SyncFluxObjectResponse,
} from "../lib/api/core/core.pb";
import { Kind } from "../lib/api/core/types.pb";
import { Automation } from "../lib/objects";
import {
  MultiRequestError,
  NoNamespace,
  ReactQueryOptions,
  RequestError,
  Syncable,
} from "../lib/types";
import { notifyError, notifySuccess } from "../lib/utils";
import { convertResponse } from "./objects";
import _ from "lodash";

type Res = {
  result: Automation[];
  errors: MultiRequestError[];
  searchedNamespaces: { [key: string]: string[] }[];
};

export type SearchedNamespaces = { [key: string]: string[] }[];

export function useListAutomations(
  namespace: string = NoNamespace,
  opts: ReactQueryOptions<Res, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<Res, RequestError>(
    ["automations", namespace],
    () => {
      const p = [Kind.HelmRelease, Kind.Kustomization].map((kind) =>
        api
          .ListObjects({ namespace, kind })
          .then((response: ListObjectsResponse) => {
            if (!response.objects) response.objects = [];
            if (!response.errors) response.errors = [];
            return { kind, response };
          })
      );
      return Promise.all(p).then((responses) => {
        const final = {
          result: [],
          errors: [],
          searchedNamespaces: [] as SearchedNamespaces,
        };
        for (const { kind, response } of responses) {
          final.result.push(
            ...response.objects.map(
              (o) => convertResponse(kind, o) as Automation
            )
          );
          final.errors.push(
            ...response.errors.map((o) => {
              return { ...o, kind };
            })
          );
          for (var k of Object.keys(response.searchedNamespaces)) {
            const existingKeys = final.searchedNamespaces.map(
              (ns) => Object.keys(ns)[0]
            );
            if (!existingKeys.includes(k)) {
              final.searchedNamespaces.push({
                [k as string]: response.searchedNamespaces[k].namespaces,
              });
            } else {
              final.searchedNamespaces[existingKeys.indexOf(k)][k] = _.uniq([
                ...final.searchedNamespaces[existingKeys.indexOf(k)][k],
                ...response.searchedNamespaces[k].namespaces,
              ]);
            }
          }
        }
        return final;
      });
    },
    opts
  );
}

export function useSyncFluxObject(objs: Syncable[]) {
  const { api } = useContext(CoreClientContext);
  const mutation = useMutation<
    SyncFluxObjectResponse,
    RequestError,
    SyncFluxObjectRequest
  >(({ withSource }) => api.SyncFluxObject({ objects: objs, withSource }), {
    onSuccess: () => notifySuccess("Sync request successful!"),
    onError: (error) => notifyError(error.message),
  });

  return mutation;
}
