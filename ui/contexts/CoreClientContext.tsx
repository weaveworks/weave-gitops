import { useQuery } from "@tanstack/react-query";
import * as React from "react";
import { Core, GetFeatureFlagsResponse } from "../lib/api/core/core.pb";
import { TokenRefreshWrapper } from "../lib/requests";
import { RequestError } from "../lib/types";
import {
  getBasePath,
  reloadBrowserSignIn,
  stripBasePath,
  withBasePath,
} from "../lib/utils";

type Props = {
  api: typeof Core;
  children: any;
};

export type FeatureFlags = { [key: string]: string };

export type CoreClientContextType = {
  api: typeof Core;
  featureFlags: FeatureFlags;
};

export const CoreClientContext =
  React.createContext<CoreClientContextType | null>(null);

function FeatureFlags(api) {
  const { data } = useQuery<GetFeatureFlagsResponse, RequestError>({
    queryKey: ["feature_flags"],
    queryFn: () => api.GetFeatureFlags({}),
    staleTime: Infinity,
    gcTime: Infinity,
  });
  return data?.flags;
}

export async function refreshToken() {
  const res = await fetch(withBasePath("/oauth2/refresh"), { method: "POST" });
  if (!res.ok) {
    // The login redirect system is aware of the base URL and will add it,
    // so we need to strip it off here.
    reloadBrowserSignIn(stripBasePath(location.pathname) + location.search);

    // Return a promse that does not resolve.
    // This stops any more API requests or refreshToken requests from being
    // made during the few seconds the browser is redirecting.
    return new Promise<void>(() => null);
  }
}

export function UnAuthorizedInterceptor(api: typeof Core): typeof Core {
  return TokenRefreshWrapper.wrap(api, refreshToken);
}

export function setAPIPathPrefix(api: any) {
  const wrapped: any = {};
  for (const method of Object.getOwnPropertyNames(api)) {
    if (typeof api[method] != "function") {
      continue;
    }
    wrapped[method] = (req: any, initReq: any) => {
      const initWithBaseURL = { pathPrefix: getBasePath(), ...initReq };
      return api[method](req, initWithBaseURL);
    };
  }
  return wrapped;
}

export default function CoreClientContextProvider({ api, children }: Props) {
  const wrapped = UnAuthorizedInterceptor(setAPIPathPrefix(api));

  return (
    <CoreClientContext.Provider
      value={{ api: wrapped, featureFlags: FeatureFlags(wrapped) }}
    >
      {children}
    </CoreClientContext.Provider>
  );
}
