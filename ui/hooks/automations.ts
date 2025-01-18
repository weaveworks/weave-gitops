import { useMutation, useQuery } from "@tanstack/react-query";
import { useContext } from "react";
import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  ListObjectsResponse,
  SyncFluxObjectRequest,
  SyncFluxObjectResponse,
} from "../lib/api/core/core.pb";
import { Kind, ObjectRef } from "../lib/api/core/types.pb";
import { Automation } from "../lib/objects";
import {
  MultiRequestError,
  NoNamespace,
  ReactQueryOptions,
  RequestError,
  SearchedNamespaces,
} from "../lib/types";
import { notifyError, notifySuccess } from "../lib/utils";
import { convertResponse } from "./objects";

type Res = {
  result: Automation[];
  errors: MultiRequestError[];
  searchedNamespaces: SearchedNamespaces;
};

export function useListAutomations(
  namespace: string = NoNamespace,
  opts: ReactQueryOptions<Res, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  },
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<Res, RequestError>({
    queryKey: ["automations", namespace],
    queryFn: () => {
      const p = [Kind.HelmRelease, Kind.Kustomization].map((kind) =>
        api
          .ListObjects({ namespace, kind })
          .then((response: ListObjectsResponse) => {
            if (!response.objects) response.objects = [];
            if (!response.errors) response.errors = [];
            return { kind, response };
          }),
      );
      return Promise.all(p).then((responses) => {
        const final: Res = {
          result: [],
          errors: [],
          searchedNamespaces: {},
        };
        for (const { kind, response } of responses) {
          final.result.push(
            ...response.objects.map(
              (o) => convertResponse(kind, o) as Automation,
            ),
          );
          final.errors.push(
            ...response.errors.map((o) => {
              return { ...o, kind };
            }),
          );
          final.searchedNamespaces[kind] = response.searchedNamespaces;
        }
        return final;
      });
    },
    ...opts,
  });
}

export function useSyncFluxObject(objs: ObjectRef[]) {
  const { api } = useContext(CoreClientContext);
  const mutation = useMutation<
    SyncFluxObjectResponse,
    RequestError,
    SyncFluxObjectRequest
  >({
    mutationFn: ({ withSource }) =>
      api.SyncFluxObject({ objects: objs, withSource }),
    onSuccess: () => notifySuccess("Sync request successful!"),
    onError: (error) => notifyError(error.message),
  });

  return mutation;
}
