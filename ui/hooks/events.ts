import { useQuery } from "@tanstack/react-query";
import { useContext } from "react";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { ListEventsResponse } from "../lib/api/core/core.pb";
import { ObjectRef } from "../lib/api/core/types.pb";
import { ReactQueryOptions, RequestError } from "../lib/types";

export function useListEvents(
  obj: ObjectRef,
  opts: ReactQueryOptions<ListEventsResponse, RequestError> = {
    retry: false,
    refetchOnWindowFocus: false,
    refetchInterval: 5000,
  },
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<ListEventsResponse, RequestError>({
    queryKey: ["events", obj],
    queryFn: () =>
      api.ListEvents({
        involvedObject: obj,
      }),
    ...opts,
  });
}
