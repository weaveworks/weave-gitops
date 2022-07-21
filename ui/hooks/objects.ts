import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { GetObjectResponse } from "../lib/api/core/core.pb";
import { Object as ResponseObject } from "../lib/api/core/types.pb";
import { FluxObject, Kind } from "../lib/objects";
import { RequestError } from "../lib/types";

function convertResponse(response: ResponseObject): FluxObject {
  const fluxObject = new FluxObject(response);
  return fluxObject;
}

export function useGetObject(
  name: string,
  namespace: string,
  kind: Kind,
  clusterName: string
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<FluxObject, RequestError>(
    ["object", clusterName, kind, namespace, name],
    () =>
      api
        .GetObject({ name, namespace, kind, clusterName })
        .then((result: GetObjectResponse) => convertResponse(result.object)),
    { retry: false, refetchInterval: 5000 }
  );
}
