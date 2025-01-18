import { useQuery } from "@tanstack/react-query";
import { useContext } from "react";

import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  ListPoliciesRequest,
  ListPoliciesResponse,
  GetPolicyRequest,
  GetPolicyResponse,
} from "../lib/api/core/core.pb";
import { ReactQueryOptions, RequestError } from "../lib/types";

const LIST_POLICIES_QUERY_KEY = "list-policy";

export function useListPolicies(
  req: ListPoliciesRequest,
  opts: ReactQueryOptions<ListPoliciesResponse, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  },
) {
  const { api } = useContext(CoreClientContext);
  return useQuery<ListPoliciesResponse, Error>({
    queryKey: [LIST_POLICIES_QUERY_KEY, req],
    queryFn: () => api.ListPolicies(req),
    ...opts,
  });
}
const GET_POLICY_QUERY_KEY = "get-policy-details";

export function useGetPolicyDetails(
  req: GetPolicyRequest,
  opts: ReactQueryOptions<GetPolicyResponse, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  },
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<GetPolicyResponse, Error>({
    queryKey: [GET_POLICY_QUERY_KEY, req],
    queryFn: () => api.GetPolicy(req),
    ...opts,
  });
}
