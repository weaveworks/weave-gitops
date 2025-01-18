import { useQuery } from "@tanstack/react-query";
import { useContext } from "react";

import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  GetPolicyValidationRequest,
  GetPolicyValidationResponse,
  ListPolicyValidationsRequest,
  ListPolicyValidationsResponse,
} from "../lib/api/core/core.pb";
import { ReactQueryOptions, RequestError } from "../lib/types";

const LIST_POLICY_VIOLATION_QUERY_KEY = "list-policy-violations";

export function useListPolicyValidations(
  req: ListPolicyValidationsRequest,
  opts: ReactQueryOptions<ListPolicyValidationsResponse, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  },
) {
  const { api } = useContext(CoreClientContext);
  return useQuery<ListPolicyValidationsResponse, Error>({
    queryKey: [LIST_POLICY_VIOLATION_QUERY_KEY, req],
    queryFn: () => api.ListPolicyValidations(req),
    ...opts,
  });
}

const GET_POLICY_VIOLATION_QUERY_KEY = "get-policy-violation-details";

export function useGetPolicyValidationDetails(
  req: GetPolicyValidationRequest,
  opts: ReactQueryOptions<GetPolicyValidationResponse, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  },
) {
  const { api } = useContext(CoreClientContext);
  return useQuery<GetPolicyValidationResponse, Error>({
    queryKey: [GET_POLICY_VIOLATION_QUERY_KEY, req],
    queryFn: () => api.GetPolicyValidation(req),
    ...opts,
  });
}
