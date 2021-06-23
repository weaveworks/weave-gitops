import { useContext, useEffect, useState } from "react";
import { AppContext } from "../contexts/AppContext";
import { Application } from "../lib/api/applications/applications.pb";

export default function useApplications() {
  const { applicationsClient } = useContext(AppContext);
  const [applications, setApplications] = useState<Application[]>([]);

  useEffect(() => {
    applicationsClient
      .ListApplications({})
      .then((res) => setApplications(res.applications));
  }, []);

  const getApplication = (applicationName: string) =>
    applicationsClient.GetApplication({ applicationName });

  return {
    applications,
    getApplication,
  };
}
