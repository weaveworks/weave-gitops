import { useContext, useState } from "react";
import { AppContext } from "../contexts/AppContext";
import { RequestError } from "../lib/types";

export default function useCommon() {
  const { appState } = useContext(AppContext);

  return { appState };
}

type RequestState<T> = {
  value: T;
  error: RequestError;
  loading: boolean;
};

export function useRequestState<T>(): [
  T,
  boolean,
  RequestError,
  (p: Promise<T>) => void
] {
  const [state, setState] = useState<RequestState<T>>({
    value: null,
    loading: true,
    error: null,
  });

  function req(p: Promise<T>) {
    setState({ ...state, loading: true });
    return p
      .then((res) => setState({ value: res, loading: false, error: null }))
      .catch((error) => setState({ error, loading: false, value: null }));
  }

  return [state.value, state.loading, state.error, req];
}
