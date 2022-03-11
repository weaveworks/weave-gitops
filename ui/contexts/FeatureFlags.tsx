import * as React from "react";

export type FeatureFlagsContext = {
  authFlag: boolean | null;
  clusterUserFlag: boolean | null;
  OIDCFlag: boolean | null;
};

export const FeatureFlags =
  React.createContext<FeatureFlagsContext | null>(null);

export default function FeatureFlagsContextProvider({ children }) {
  const [authFlag, setAuthFlag] = React.useState<boolean>(null);
  const [clusterUserFlag, setClusterUserFlag] = React.useState<boolean>(null);
  const [OIDCFlag, setOIDCFlag] = React.useState<boolean>(null);

  const getAuthFlag = React.useCallback(() => {
    fetch("/v1/featureflags")
      .then((response) => response.json())
      .then((data) => {
        setAuthFlag(data.flags.WEAVE_GITOPS_AUTH_ENABLED === "true");
        setClusterUserFlag(data.flags.CLUSTER_USER_AUTH === "true");
        setOIDCFlag(data.flags.OIDC_AUTH === "true");
      })
      .catch((err) => console.log(err));
  }, []);

  React.useEffect(() => {
    getAuthFlag();
  }, [getAuthFlag]);

  // Loading...
  if (authFlag === null) {
    return null;
  }

  return (
    <FeatureFlags.Provider
      value={{
        authFlag,
        clusterUserFlag,
        OIDCFlag,
      }}
    >
      {children}
    </FeatureFlags.Provider>
  );
}
