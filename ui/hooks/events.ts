import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { ListEventsResponse } from "../lib/api/core/core.pb";
import { ObjectRef } from "../lib/api/core/types.pb";
import { ReactQueryOptions, RequestError } from "../lib/types";
import { removeKind } from "../lib/utils";

export function useListEvents(
  obj: ObjectRef,
  opts: ReactQueryOptions<ListEventsResponse, RequestError> = {
    retry: false,
    refetchOnWindowFocus: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<ListEventsResponse, RequestError>(
    ["events", obj],
    () =>
      api.ListEvents({
        involvedObject: { ...obj, kind: removeKind(obj.kind) },
      }),
    opts
  );
}
