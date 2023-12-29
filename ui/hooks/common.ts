import { useContext, useEffect, useState } from "react";
import { AppContext } from "../contexts/AppContext";

export default function useCommon() {
  const { appState, settings } = useContext(AppContext);

  return { appState, settings };
}

// Copied and TS-ified from https://usehooks.com/useDebounce/
export function useDebounce<T>(value: T, delay: number) {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);
    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  if (process.env.NODE_ENV === "test") {
    return value;
  }

  return debouncedValue;
}
