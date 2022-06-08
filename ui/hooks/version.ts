import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { GetVersionResponse } from "../lib/api/core/core.pb";
import { RequestError } from "../lib/types";

export function useVersion() {
  const { api } = useContext(CoreClientContext);

  return useQuery<GetVersionResponse, RequestError>(
    "version",
    () => api.GetVersion({}),
    {
      staleTime: Infinity,
      cacheTime: Infinity,
    }
  );
}
