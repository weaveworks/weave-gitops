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
  data: VersionType;
  loading: boolean;
  error: Error;
};

export const VersionContext =
  React.createContext<VersionContextType | null>(null);

export default function VersionContextProvider({ children }: Props) {
  const { request } = React.useContext(AppContext);
  const [data, setData] = React.useState(null);
  const [loading, setLoading] = React.useState(null);
  const [error, setError] = React.useState(null);

  React.useEffect(() => {
    setLoading(true);
    request("/v1/version")
      .then((response) => response.json())
      .then((data) => setData(data))
      .catch((error) => setError(error))
      .finally(() => setLoading(false));
  }, []);

  return (
    <VersionContext.Provider value={{ data: data?.version, loading, error }}>
      {children}
    </VersionContext.Provider>
  );
}
