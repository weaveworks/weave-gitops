import { useContext } from "react";
import { AppContext } from "../contexts/AppContext";

export default function useCommon() {
  const { appState } = useContext(AppContext);

  return { appState };
}
