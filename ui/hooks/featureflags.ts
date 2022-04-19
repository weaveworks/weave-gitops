import { useContext } from "react";
import { FeatureFlags, Flags } from "../contexts/FeatureFlags";

export function useFeatureFlags() {
  const { flags } = useContext(FeatureFlags);

  return flags || ({} as Flags);
}
