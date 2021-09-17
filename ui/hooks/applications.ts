import { useContext, useState } from "react";
import { AppContext } from "../contexts/AppContext";
import { Application } from "../lib/api/applications/applications.pb";

const WeGONamespace = "wego-system";

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
