import { useQuery } from "react-query";
import { RequestError } from "../lib/types";

// Taken straight from the TS docs:
// https://www.typescriptlang.org/docs/handbook/2/mapped-types.html
type OptionsFlags<Type> = {
  [Property in keyof Type]: boolean;
};

type FeatureFlags = {
  CLUSTER_USER_AUTH: () => void;
  OIDC_AUTH: () => void;
};

export type Flags = OptionsFlags<FeatureFlags>;

export function useFeatureFlags() {
  const getFlags = () =>
    fetch("/v1/featureflags").then((response) => response.json());

  return useQuery<{ flags: Flags }, RequestError>(
    "feature_flags",
    () => getFlags(),
    {
      staleTime: Infinity,
      cacheTime: Infinity,
    }
  );
}
