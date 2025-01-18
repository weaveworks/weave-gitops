import { useQuery } from "@tanstack/react-query";
import { useContext } from "react";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { GetVersionResponse } from "../lib/api/core/core.pb";
import { RequestError } from "../lib/types";

export function useVersion() {
  const { api } = useContext(CoreClientContext);

  return useQuery<GetVersionResponse, RequestError>({
    queryKey: ["version"],
    queryFn: () => api.GetVersion({}),
    staleTime: Infinity,
    gcTime: Infinity,
  });
}
