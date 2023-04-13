import qs from "query-string";
import * as React from "react";
import { useQuery } from "react-query";
import { Core, GetFeatureFlagsResponse } from "../lib/api/core/core.pb";
import { RequestError } from "../lib/types";
import { TokenRefreshWrapper } from "../lib/requests";
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
  const res = await fetch("/oauth2/refresh", { method: "POST" });
  if (!res.ok) {
    window.location.replace(
      AuthRoutes.AUTH_PATH_SIGNIN +
        "?" +
        qs.stringify({
          redirect: location.pathname + location.search,
        })
    );
  }
}

export function UnAuthorizedInterceptor(api: typeof Core): typeof Core {
  return TokenRefreshWrapper.wrap(api, refreshToken);
}

export default function CoreClientContextProvider({ api, children }: Props) {
  const wrapped = UnAuthorizedInterceptor(api);

  return (
    <CoreClientContext.Provider
      value={{ api: wrapped, featureFlags: FeatureFlags(wrapped) }}
    >
      {children}
    </CoreClientContext.Provider>
  );
}
