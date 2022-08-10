import { useContext } from "react";
import { useQuery } from "react-query";
import { RequestError } from "../lib/types";
import { GetFeatureFlagsResponse } from "../lib/api/core/core.pb";
import { CoreClientContext } from "../contexts/CoreClientContext";

// Taken straight from the TS docs:
// https://www.typescriptlang.org/docs/handbook/2/mapped-types.html
type OptionsFlags<Type> = {
  [Property in keyof Type]: boolean;
};

type FeatureFlags = {
  CLUSTER_USER_AUTH: () => void;
  OIDC_AUTH: () => void;
  WEAVE_GITOPS_FEATURE_TENANCY: () => void;
};

export type Flags = OptionsFlags<FeatureFlags>;

export function useFeatureFlags() {
  const { api } = useContext(CoreClientContext);
  return useQuery<GetFeatureFlagsResponse, RequestError>(
    "feature_flags",
    () => api.GetFeatureFlags({}),
    {
      staleTime: Infinity,
      cacheTime: Infinity,
    }
  );
}
