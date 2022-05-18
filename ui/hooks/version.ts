import { useContext } from "react";
import { useQuery } from "react-query";
import { RequestError } from "../lib/types";
import { GetVersionResponse } from "../lib/api/core/core.pb";
import { CoreClientContext } from "../contexts/CoreClientContext";

export type Version = {
  version: string;
  gitCommit: string;
  branch: string;
  buildTime: string;
};

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
