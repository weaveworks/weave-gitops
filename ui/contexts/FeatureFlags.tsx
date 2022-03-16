import * as React from "react";
import { AppContext } from "./AppContext";

// Taken straight from the TS docs:
// https://www.typescriptlang.org/docs/handbook/2/mapped-types.html
type OptionsFlags<Type> = {
  [Property in keyof Type]: boolean;
};

type FeatureFlags = {
  WEAVE_GITOPS_AUTH_ENABLED: () => void;
  CLUSTER_USER_AUTH: () => void;
  OIDC_AUTH: () => void;
};

export type Flags = OptionsFlags<FeatureFlags>;

export type FeatureFlagsContextType = Flags;

export const FeatureFlags =
  React.createContext<FeatureFlagsContextType | null>(null);

export default function FeatureFlagsContextProvider({ children }) {
  const { request } = React.useContext(AppContext);
  const [data, setData] = React.useState(null);

  React.useEffect(() => {
    request("/v1/featureflags")
      .then((response) => response.json())
      .then((data) => setData(data));
  }, []);

  return (
    <FeatureFlags.Provider value={data?.flags}>
      {children}
    </FeatureFlags.Provider>
  );
}
