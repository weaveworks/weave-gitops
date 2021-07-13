import { useContext, useState } from "react";
import { AppContext } from "../contexts/AppContext";
import { Application } from "../lib/api/applications/applications.pb";
import { AsyncError } from "../lib/types";

const WeGONamespace = "wego-system";

export default function useApplications() {
  const { applicationsClient } = useContext(AppContext);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<AsyncError>(null);
  const [applications, setApplications] = useState<Application[]>([]);
  const [currentApplication, setCurrentApplication] = useState<Application>({});

  const getApplication = (name: string) => {
    applicationsClient
      .GetApplication({ name, namespace: WeGONamespace })
      .then(({ application }) => setCurrentApplication(application))
      .catch((err) =>
        setError({ message: "Could not get application", detail: err.message })
      )
      .finally(() => setLoading(false));
  };

  const listApplications = () => {
    applicationsClient
      .ListApplications({})
      .then(({ applications }) => setApplications(applications))
      .catch((err) => setError(err))
      .finally(() => setLoading(false));
  };

  return {
    loading,
    error,
    currentApplication,
    applications,
    getApplication,
    listApplications,
  };
}
