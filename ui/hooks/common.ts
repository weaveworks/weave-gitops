import { useContext } from "react";
import { AppContext } from "../contexts/AppContext";

export default function useCommon() {
  const { appState, settings } = useContext(AppContext);

  return { appState, settings };
}
