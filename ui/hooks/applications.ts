import { useContext, useEffect, useState } from "react";
import { AppContext } from "../contexts/AppContext";
import { Application } from "../lib/api/applications/applications.pb";

export default function useApplications() {
  const { applicationsClient, doAsyncError } = useContext(AppContext);
  const [loading, setLoading] = useState(true);
  const [applications, setApplications] = useState<Application[]>([]);

  useEffect(() => {
    setLoading(true);

    applicationsClient
      .ListApplications({})
      .then((res) => setApplications(res.applications))
      .catch((err) => {
        doAsyncError(err.message, err.detail);
      })
      .finally(() => setLoading(false));
  }, []);

  const getApplication = (applicationName: string) => {
    setLoading(true);

    return applicationsClient
      .GetApplication({ applicationName })
      .then((res) => res.application)
      .catch((err) => doAsyncError("Error fetching application", err.message))
      .finally(() => setLoading(false));
  };

  return {
    loading,
    applications,
    getApplication,
  };
}
