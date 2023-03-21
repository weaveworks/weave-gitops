import { useContext } from "react";
import { CoreClientContext } from "../contexts/CoreClientContext";

export type FeatureFlags = { [key: string]: string };

export function useFeatureFlags() {
  const { featureFlags } = useContext(CoreClientContext);

  const isFlagEnabled = (flag: string) =>
    featureFlags?.["WEAVE_GITOPS_FEATURE_OIDC_BUTTON_LABEL"] !== "" ||
    featureFlags?.[flag] === "true";

  return { isFlagEnabled };
}
