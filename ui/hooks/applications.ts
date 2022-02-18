import _ from "lodash";
import { AppContext } from "../contexts/AppContext";
import {
  GitProvider,
  ListCommitsRequest,
  ListCommitsResponse,
  ParseRepoURLResponse,
} from "../lib/api/applications/applications.pb";
import { WeGONamespace } from "../lib/types";
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
