import { useContext } from "react";
import { CoreClientContext } from "../contexts/CoreClientContext";

export type FeatureFlags = { [key: string]: string };

export function useFeatureFlags() {
  const { featureFlags } = useContext(CoreClientContext);

  const isFlagEnabled = (flag: string) => featureFlags?.[flag] === "true";

  return { isFlagEnabled };
}
