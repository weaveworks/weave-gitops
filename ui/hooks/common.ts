import { useContext, useEffect, useState } from "react";
import { AppContext } from "../contexts/AppContext";
import { GitProvider } from "../lib/api/gitauth/gitauth.pb";
import { RequestError } from "../lib/types";

export default function useCommon() {
  const { appState, settings } = useContext(AppContext);

  return { appState, settings };
}

export type RequestState<T> = {
  value: T;
  error: RequestError;
  loading: boolean;
};

type RequestWithToken<T> = (provider: GitProvider, body: T) => void;

export type RequestStateWithToken<Req, Res> = [
  res: Res,
  loading: boolean,
  error: RequestError,
  req: RequestWithToken<Req>
];

export type ReturnType<T> = [T, boolean, RequestError, (p: Promise<T>) => void];

export function useRequestState<T>(): ReturnType<T> {
  const [state, setState] = useState<RequestState<T>>({
    value: null,
    loading: false,
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

// Copied and TS-ified from https://usehooks.com/useDebounce/
export function useDebounce<T>(value: T, delay: number) {
  if (process.env.NODE_ENV === "test") {
    return value;
  }

  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);
    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
}

const providerTokenHeaderName = "Git-Provider-Token";

export function makeHeaders(tokenGetter: () => string) {
  const token = tokenGetter();

  return new Headers({
    [providerTokenHeaderName]: `token ${token}`,
  });
}
