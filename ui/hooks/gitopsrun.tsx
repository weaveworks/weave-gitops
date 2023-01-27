import { useContext } from "react";
import { useQuery } from "react-query";
import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  GetSessionLogsRequest,
  GetSessionLogsResponse,
} from "../lib/api/core/core.pb";

export const useGetLogs = (req: GetSessionLogsRequest) => {
  const { api } = useContext(CoreClientContext);

  console.log("calling get session logs");

  const { isLoading, data, error } = useQuery<GetSessionLogsResponse, Error>(
    `logs ${req.sessionId} ${req.clusterName} ${req.namespace} ${req.token}`,
    () => api.GetSessionLogs(req)
  );
  return { isLoading, data, error };
};
