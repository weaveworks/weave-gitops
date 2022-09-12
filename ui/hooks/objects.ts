import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { GetObjectResponse, ListError } from "../lib/api/core/core.pb";
import { Object as ResponseObject } from "../lib/api/core/types.pb";
import {
  Alert,
  Bucket,
  FluxObject,
  GitRepository,
  HelmChart,
  HelmRepository,
  OCIRepository,
  Kustomization,
  HelmRelease,
  Kind,
  Provider,
} from "../lib/objects";
import { ReactQueryOptions, RequestError } from "../lib/types";

export function convertResponse(kind: Kind, response?: ResponseObject) {
  if (kind === Kind.HelmRepository) {
    return new HelmRepository(response);
  }
  if (kind === Kind.HelmChart) {
    return new HelmChart(response);
  }
  if (kind === Kind.Bucket) {
    return new Bucket(response);
  }
  if (kind === Kind.GitRepository) {
    return new GitRepository(response);
  }
  if (kind === Kind.OCIRepository) {
    return new OCIRepository(response);
  }
  if (kind === Kind.Kustomization) {
    return new Kustomization(response);
  }
  if (kind === Kind.HelmRelease) {
    return new HelmRelease(response);
  }
  if (kind === Kind.Provider) {
    return new Provider(response);
  }
  if (kind === Kind.Alert) {
    return new Alert(response);
  }

  return new FluxObject(response);
}

export function useGetObject<T extends FluxObject>(
  name: string,
  namespace: string,
  kind: Kind,
  clusterName: string,
  opts: ReactQueryOptions<T, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);

  const response = useQuery<T, RequestError>(
    ["object", clusterName, kind, namespace, name],
    () =>
      api
        .GetObject({ name, namespace, kind, clusterName })
        .then(
          (result: GetObjectResponse) =>
            convertResponse(kind, result.object) as T
        ),
    opts
  );
  if (response.error) {
    return { ...response, data: convertResponse(kind) as T };
  }
  return response;
}

type Res = { objects: FluxObject[]; errors: ListError[] };

export function useListObjects<T extends FluxObject>(
  namespace: string,
  kind: Kind,
  opts: ReactQueryOptions<Res, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<Res, RequestError>(
    ["objects", namespace],
    () => {
      return api.ListObjects({ namespace, kind }).then((res) => {
        const objects = res.objects?.map(
          (obj) => convertResponse(kind, obj) as T
        );
        return { objects: objects, errors: res.errors };
      });
    },
    opts
  );
}
