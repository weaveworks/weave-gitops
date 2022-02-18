import _ from "lodash";
import { useContext, useEffect } from "react";
import { AppContext } from "../contexts/AppContext";
import {
  GitProvider,
  ListCommitsRequest,
  ListCommitsResponse,
  ParseRepoURLResponse,
} from "../lib/api/applications/applications.pb";
import { makeHeaders, RequestStateWithToken, useRequestState } from "./common";

export function useParseRepoURL(url: string) {
  const { applicationsClient } = useContext(AppContext);
  const [res, loading, error, req] = useRequestState<ParseRepoURLResponse>();

  useEffect(() => {
    req(applicationsClient.ParseRepoURL({ url }));
  }, [url]);

  return [res, loading, error];
}

type ListCommitsReturnType = RequestStateWithToken<
  ListCommitsRequest,
  ListCommitsResponse
>;

export function useListCommits(): ListCommitsReturnType {
  const { applicationsClient } = useContext(AppContext);
  const [res, loading, error, req] = useRequestState<ListCommitsResponse>();

  return [
    res,
    loading,
    error,
    (provider: GitProvider, body: ListCommitsRequest) => {
      const headers = makeHeaders(_.bind(_, this, provider));
      req(applicationsClient.ListCommits(body, { headers }));
    },
  ];
}

export default function useApplications() {
  const { applicationsClient } = useContext(AppContext);

  const listCommits = (app: any) => {
    return applicationsClient.ListCommits({ ...app });
  };

  return {
    listCommits,
  };
}
