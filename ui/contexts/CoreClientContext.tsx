import qs from "query-string";
import * as React from "react";
import { Core } from "../lib/api/core/core.pb";
import { AuthRoutes } from "./AuthContext";

type Props = {
  api: typeof Core;
  children: any;
};

export type CoreClientContextType = {
  api: typeof Core;
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

export default function CoreClientContextProvider({ api, children }: Props) {
  const wrapped = UnAuthorizedInterceptor(api) as typeof Core;

  return (
    <CoreClientContext.Provider value={{ api: wrapped }}>
      {children}
    </CoreClientContext.Provider>
  );
}
