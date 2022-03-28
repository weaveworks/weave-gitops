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
    { retry: false, refetchInterval: 5000 }
  );
}

export function useGetReconciledObjects(
  name: string,
  namespace: string,
  type: AutomationKind,
  kinds: GroupVersionKind[]
) {
  const { api } = useContext(AppContext);
  const [res, setRes] = useState({ data: null, isLoading: false, error: null });

  useEffect(() => {
    if (!name || !namespace || !type || kinds.length === 0) {
      return;
    }

    setRes({ ...res, isLoading: true });

    getChildren(api, name, namespace, kinds)
      .then((res) => setRes({ data: res, isLoading: false, error: null }))
      .catch((e) => setRes({ data: null, error: e, isLoading: false }));
  }, [type, kinds]);

  return res;
}
