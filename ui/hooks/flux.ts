import { useContext } from "react";
import { useQuery } from "react-query";
import { AppContext } from "../contexts/AppContext";
import { ListFluxRuntimeObjectsResponse } from "../lib/api/core/core.pb";
import { RequestError, WeGONamespace } from "../lib/types";

export function useListFluxRuntimeObjects(namespace = WeGONamespace) {
  const { api } = useContext(AppContext);

  return useQuery<ListFluxRuntimeObjectsResponse, RequestError>(
    "flux_runtime_objects",
    () => api.ListFluxRuntimeObjects({ namespace }),
    { retry: false }
  );
}
