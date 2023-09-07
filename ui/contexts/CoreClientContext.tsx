import qs from "query-string";
import * as React from "react";
import { useQuery } from "react-query";
import { Core, GetFeatureFlagsResponse } from "../lib/api/core/core.pb";
import { TokenRefreshWrapper } from "../lib/requests";
import { RequestError } from "../lib/types";
import { getBaseURL, stripBaseURL, withBaseURL } from "../lib/utils";
import { AuthRoutes } from "./AuthContext";

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
  const { data } = useQuery<GetFeatureFlagsResponse, RequestError>(
    "feature_flags",
    () => api.GetFeatureFlags({}),
    {
      staleTime: Infinity,
      cacheTime: Infinity,
    }
  );
  return data?.flags;
}

export async function refreshToken() {
  const res = await fetch(withBaseURL("/oauth2/refresh"), { method: "POST" });
  if (!res.ok) {
    window.location.replace(
      withBaseURL(AuthRoutes.AUTH_PATH_SIGNIN) +
        "?" +
        qs.stringify({
          redirect: stripBaseURL(location.pathname + location.search),
        })
    );

    // Return a promse that does not resolve.
    // This stops any more API requests or refreshToken requests from being
    // made during the few seconds the browser is redirecting.
    return new Promise<void>(() => null);
  }
}

export function UnAuthorizedInterceptor(api: typeof Core): typeof Core {
  return TokenRefreshWrapper.wrap(api, refreshToken);
}

export function setBaseURL(api: any) {
  const wrapped: any = {};
  for (const method of Object.getOwnPropertyNames(api)) {
    if (typeof api[method] != "function") {
      continue;
    }
    wrapped[method] = (req: any, initReq: any) => {
      const initWithBaseURL = { pathPrefix: getBaseURL(), ...initReq };
      return api[method](req, initWithBaseURL);
    };
  }
  return wrapped;
}

export default function CoreClientContextProvider({ api, children }: Props) {
  let wrapped = UnAuthorizedInterceptor(api);
  wrapped = setBaseURL(api);

  return (
    <CoreClientContext.Provider
      value={{ api: wrapped, featureFlags: FeatureFlags(wrapped) }}
    >
      {children}
    </CoreClientContext.Provider>
  );
}
