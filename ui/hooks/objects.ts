import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { GetObjectResponse } from "../lib/api/core/core.pb";
import { Object as ResponseObject } from "../lib/api/core/types.pb";
import {
  Bucket,
  FluxObject,
  HelmChart,
  GitRepository,
  HelmRepository,
  Kind,
} from "../lib/objects";
import { RequestError } from "../lib/types";

function convertResponse(kind: Kind, response: ResponseObject) {
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

  return new FluxObject(response);
}

export function useGetObject<T extends FluxObject>(
  name: string,
  namespace: string,
  kind: Kind,
  clusterName: string
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<T, RequestError>(
    ["object", clusterName, kind, namespace, name],
    () =>
      api
        .GetObject({ name, namespace, kind, clusterName })
        .then(
          (result: GetObjectResponse) =>
            convertResponse(kind, result.object) as T
        ),
    { retry: false, refetchInterval: 5000 }
  );
}
