import * as React from "react";

export type FeatureFlagsContext = {
  authFlag: boolean | null;
};

export const FeatureFlags =
  React.createContext<FeatureFlagsContext | null>(null);

export default function FeatureFlagsContextProvider({ children }) {
  const [authFlag, setAuthFlag] = React.useState<boolean>(null);

  const getAuthFlag = React.useCallback(() => {
    fetch("/v1/featureflags")
      .then((response) => response.json())
      .then((data) =>
        setAuthFlag(data.flags.WEAVE_GITOPS_AUTH_ENABLED === "true")
      )
      .catch((err) => console.log(err));
  }, []);

  React.useEffect(() => {
    getAuthFlag();
  }, [getAuthFlag]);

  return (
    <FeatureFlags.Provider
      value={{
        authFlag,
      }}
    >
      {children}
    </FeatureFlags.Provider>
  );
}
