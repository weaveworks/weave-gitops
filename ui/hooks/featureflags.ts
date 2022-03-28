import { useContext } from "react";
import { FeatureFlags, Flags } from "../contexts/FeatureFlags";

export function useFeatureFlags() {
  const data = useContext(FeatureFlags);

  return data || ({} as Flags);
}
