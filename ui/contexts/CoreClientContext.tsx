import * as React from "react";
import { Redirect, useLocation } from "react-router-dom";
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

export function UnAuthorizedInterceptor(api: any, location: any) {
  const wrapped = {} as any;
  //   Wrap each API method in a check that redirects to the signin page if a 401 is returned.
  for (const method of Object.getOwnPropertyNames(api)) {
    if (typeof api[method] != "function") {
      continue;
    }
    wrapped[method] = (req, initReq) => {
      return api[method](req, initReq).catch((err) => {
        if (err.code === 401) {
          return (
            <Redirect
              to={{
                path: AuthRoutes.AUTH_PATH_SIGNIN,
                state: { from: location },
              }}
            />
          );
        }
        throw err;
      });
    };
  }
  return wrapped;
}

export default function CoreClientContextProvider({ api, children }: Props) {
  const location = useLocation();
  const wrapped = UnAuthorizedInterceptor(api, location) as typeof Core;

  return (
    <CoreClientContext.Provider value={{ api: wrapped }}>
      {children}
    </CoreClientContext.Provider>
  );
}
