import { useContext } from "react";
import { useQuery } from "react-query";

import { CoreClientContext } from "../contexts/CoreClientContext";
import { ReactQueryOptions, RequestError } from "../lib/types";
import {
  ListPoliciesRequest,
  ListPoliciesResponse,
  GetPolicyRequest,
  GetPolicyResponse,
} from "../lib/api/core/core.pb";

const LIST_POLICIES_QUERY_KEY = "list-policy";

export function useListPolicies(
  req: ListPoliciesRequest,
  opts: ReactQueryOptions<ListPoliciesResponse, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);
  return useQuery<ListPoliciesResponse, Error>(
    [LIST_POLICIES_QUERY_KEY, req],
    () => api.ListPolicies(req),
    opts
  );
}
const GET_POLICY_QUERY_KEY = "get-policy-details";

export function useGetPolicyDetails(
  req: GetPolicyRequest,
  opts: ReactQueryOptions<GetPolicyResponse, RequestError> = {
    retry: false,
    refetchInterval: 5000,
  }
) {
  const { api } = useContext(CoreClientContext);

  return useQuery<GetPolicyResponse, Error>(
    [GET_POLICY_QUERY_KEY, req],
    () => api.GetPolicy(req),
    opts
  );
}
