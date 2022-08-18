import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { GetObjectResponse } from "../lib/api/core/core.pb";
import { Object as ResponseObject } from "../lib/api/core/types.pb";
import {
  Bucket,
  FluxObject,
  GitRepository,
  HelmChart,
  HelmRepository,
  OCIRepository,
  Kind,
} from "../lib/objects";
import { ReactQueryOptions, RequestError } from "../lib/types";

function convertResponse(kind: Kind, response?: ResponseObject) {
  if (kind == Kind.HelmRepository) {
    return new HelmRepository(response);
  }
  if (kind == Kind.HelmChart) {
    return new HelmChart(response);
  }
  if (kind == Kind.Bucket) {
    return new Bucket(response);
  }
  if (kind == Kind.GitRepository) {
    return new GitRepository(response);
  }
  if (kind == Kind.OCIRepository) {
    return new OCIRepository(response);
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
