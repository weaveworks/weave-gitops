import { useContext } from "react";
import { VersionContext, VersionType } from "../contexts/VersionContext";

export function useVersion() {
  const { data } = useContext(VersionContext);

  return data || ({} as VersionType);
}
