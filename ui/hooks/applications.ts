import _ from "lodash";
import { useContext, useEffect, useState } from "react";
import { getProviderToken } from "..";
import { AppContext } from "../contexts/AppContext";
import {
  AddApplicationRequest,
  AddApplicationResponse,
  Application,
  GetApplicationRequest,
  GetApplicationResponse,
  GitProvider,
  ListCommitsRequest,
  ListCommitsResponse,
  ParseRepoURLResponse,
  RemoveApplicationRequest,
  RemoveApplicationResponse,
} from "../lib/api/applications/applications.pb";
import { makeHeaders, RequestStateWithToken, useRequestState } from "./common";

const WeGONamespace = "wego-system";

export function useParseRepoURL(url: string) {
  const { applicationsClient } = useContext(AppContext);
  const [res, loading, error, req] = useRequestState<ParseRepoURLResponse>();

  useEffect(() => {
    req(applicationsClient.ParseRepoURL({ url }));
  }, [url]);

  return [res, loading, error];
}

type AddApplicationReturnType = RequestStateWithToken<
  AddApplicationRequest,
  AddApplicationResponse
>;

export function useAddApplication(): AddApplicationReturnType {
  const { applicationsClient } = useContext(AppContext);
  const [res, loading, error, req] = useRequestState<AddApplicationResponse>();

  return [
    res,
    loading,
    error,
    (provider: GitProvider, body: AddApplicationRequest) => {
      const headers = makeHeaders(_.bind(getProviderToken, this, provider));
      req(applicationsClient.AddApplication(body, { headers }));
    },
  ];
}

export function useAppRemove(): RequestStateWithToken<
  RemoveApplicationRequest,
  RemoveApplicationResponse
> {
  const { applicationsClient } = useContext(AppContext);
  const [res, loading, error, req] = useRequestState<AddApplicationResponse>();

  return [
    res,
    loading,
    error,
    (provider: GitProvider, body: AddApplicationRequest) => {
      const headers = makeHeaders(_.bind(getProviderToken, this, provider));
      req(applicationsClient.RemoveApplication(body, { headers }));
    },
  ];
}

type ListCommitsReturnType = RequestStateWithToken<
  ListCommitsRequest,
  ListCommitsResponse
>;

export function useListCommits(): ListCommitsReturnType {
  const { applicationsClient, getProviderToken } = useContext(AppContext);
  const [res, loading, error, req] = useRequestState<ListCommitsResponse>();

  return [
    res,
    loading,
    error,
    (provider: GitProvider, body: ListCommitsRequest) => {
      const headers = makeHeaders(_.bind(getProviderToken, this, provider));
      req(applicationsClient.ListCommits(body, { headers }));
    },
  ];
}

export function useAppGet() {
  const { applicationsClient } = useContext(AppContext);
  const [res, loading, error, req] = useRequestState<GetApplicationResponse>();

  return [
    res,
    loading,
    error,
    (body: GetApplicationRequest) =>
      req(applicationsClient.GetApplication(body)),
  ];
}

export default function useApplications() {
  const { applicationsClient, doAsyncError } = useContext(AppContext);
  const [loading, setLoading] = useState(true);

  const listApplications = (namespace: string = WeGONamespace) => {
    setLoading(true);

    return applicationsClient
      .ListApplications({ namespace: namespace })
      .then((res) => res.applications)
      .catch((err) => doAsyncError(err.message, err.detail))
      .finally(() => setLoading(false));
  };

  const getApplication = (name: string) => {
    setLoading(true);

    return applicationsClient
      .GetApplication({ name, namespace: WeGONamespace })
      .then((res) => res.application)
      .catch((err) => doAsyncError("Error fetching application", err.message))
      .finally(() => setLoading(false));
  };

  const listCommits = (app: Application) => {
    return applicationsClient.ListCommits({ ...app });
  };

  return {
    loading,
    listApplications,
    listCommits,
    getApplication,
  };
}
