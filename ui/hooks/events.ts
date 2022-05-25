import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { ListFluxEventsResponse } from "../lib/api/core/core.pb";
import { ObjectRef } from "../lib/api/core/types.pb";
import { RequestError } from "../lib/types";

export function useListFluxEvents(obj: ObjectRef) {
  const { api } = useContext(CoreClientContext);

  return useQuery<ListFluxEventsResponse, RequestError>(
    ["events", obj],
    () =>
      api.ListFluxEvents({
        involvedObject: obj,
      }),
    {
      retry: false,
      refetchOnWindowFocus: false,
      refetchInterval: 5000,
    }
  );
}
