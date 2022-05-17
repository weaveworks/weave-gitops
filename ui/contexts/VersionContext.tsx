import * as React from "react";
import { AppContext } from "./AppContext";

type Props = {
  children?: any;
};

export type VersionType = {
  version: string;
  gitCommit: string;
  branch: string;
  buildTime: string;
};

export type VersionContextType = {
  data?: VersionType;
  loading: boolean;
  error: Error;
};

export const VersionContext =
  React.createContext<VersionContextType | null>(null);

export default function VersionContextProvider({ children }: Props) {
  const { request } = React.useContext(AppContext);
  const [data, setData] = React.useState(null);
  const [loading, setLoading] = React.useState<boolean>(true);
  const [error, setError] = React.useState(null);

  React.useEffect(() => {
    setLoading(true);
    request("/v1/version")
      .then((response) => response.json())
      .then((data) => setData(data))
      .catch((error) => setError(error))
      .finally(() => setLoading(false));
  }, []);

  const version = data?.version;
  const value: VersionType = version
    &&
    {
      version: version.version,
      gitCommit: version["git-commit"],
      branch: version.branch,
      buildTime: version["buildtime"]
    };

  return (
    <VersionContext.Provider value={{ data: value, loading, error }}>
      {children}
    </VersionContext.Provider>
  );
}
