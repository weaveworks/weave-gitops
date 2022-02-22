import { useContext, useEffect, useState } from "react";
import { useQuery } from "react-query";
import { AppContext } from "../contexts/AppContext";
import { ListFluxRuntimeObjectsResponse } from "../lib/api/core/core.pb";
import { AutomationKind, GroupVersionKind } from "../lib/api/core/types.pb";
import { getChildren } from "../lib/graph";
import { RequestError, WeGONamespace } from "../lib/types";

export function useListFluxRuntimeObjects(namespace = WeGONamespace) {
  const { api } = useContext(AppContext);

  return useQuery<ListFluxRuntimeObjectsResponse, RequestError>(
    "flux_runtime_objects",
    () => api.ListFluxRuntimeObjects({ namespace }),
    { retry: false }
  );
}

export function useGetReconciledObjects(
  name: string,
  namespace: string,
  type: AutomationKind,
  kinds: GroupVersionKind[]
) {
  const { api } = useContext(AppContext);
  const [res, setRes] = useState([]);

  useEffect(() => {
    if (!name || !namespace || !type || kinds.length === 0) {
      return;
    }

    getChildren(api, name, namespace, kinds).then((res) => setRes(res));
  }, [type, kinds]);

  return res;
}
