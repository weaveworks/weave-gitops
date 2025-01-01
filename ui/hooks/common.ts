import { useContext, useEffect, useState } from "react";
import { AppContext } from "../contexts/AppContext";
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

export type ReturnType<T> = [T, boolean, RequestError, (p: Promise<T>) => void];

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
