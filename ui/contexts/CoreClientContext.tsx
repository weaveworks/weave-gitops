import qs from "query-string";
import * as React from "react";
import { useQuery } from "react-query";
import { Core, GetFeatureFlagsResponse } from "../lib/api/core/core.pb";
import { RequestError } from "../lib/types";
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

export function UnAuthorizedInterceptor(api: any) {
  const wrapped = {} as any;
  //   Wrap each API method in a check that redirects to the signin page if a 401 is returned.
  for (const method of Object.getOwnPropertyNames(api)) {
    if (typeof api[method] != "function") {
      continue;
    }
    wrapped[method] = (req, initReq) => {
      return api[method](req, initReq).catch((err) => {
        if (err.code === 401) {
          window.location.replace(
            AuthRoutes.AUTH_PATH_SIGNIN +
              "?" +
              qs.stringify({
                redirect: location.pathname + location.search,
              })
          );
        }
        throw err;
      });
    };
  }
  return wrapped;
}

function FeatureFlags(api) {
  const { data } = useQuery<GetFeatureFlagsResponse, RequestError>(
    "feature_flags",
    () => api.GetFeatureFlags({}),
    {
      staleTime: Infinity,
      cacheTime: Infinity,
    }
  );
  return data?.flags || {};
}

export default function CoreClientContextProvider({ api, children }: Props) {
  const wrapped = UnAuthorizedInterceptor(api) as typeof Core;

  return (
    <CoreClientContext.Provider
      value={{ api: wrapped, featureFlags: FeatureFlags(wrapped) }}
    >
      {children}
    </CoreClientContext.Provider>
  );
}
