import { useContext, useEffect, useState } from "react";
import { AppContext } from "../contexts/AppContext";
import {
  Application,
  ParseRepoURLResponse,
} from "../lib/api/applications/applications.pb";
import { useRequestState } from "./common";

const WeGONamespace = "wego-system";

export function useParseRepoURL(url: string) {
  const { applicationsClient } = useContext(AppContext);
  const [res, loading, error, req] = useRequestState<ParseRepoURLResponse>();

  useEffect(() => {
    console.log("running");
    req(applicationsClient.ParseRepoURL({ url }));
  }, [url]);

  return [res, loading, error];
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
