import { useContext } from "react";
import { CoreClientContext } from "../contexts/CoreClientContext";

export type FeatureFlags = { [key: string]: string };

export function useFeatureFlags() {
  const { featureFlags } = useContext(CoreClientContext);

  const isFlagEnabled = (flag: string) => {
    if (flag === "WEAVE_GITOPS_FEATURE_OIDC_BUTTON_LABEL") {
      return featureFlags?.["WEAVE_GITOPS_FEATURE_OIDC_BUTTON_LABEL"] !== "";
    } else return featureFlags?.[flag] === "true";
  };

  return { isFlagEnabled };
}
