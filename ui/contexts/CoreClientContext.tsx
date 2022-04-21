import _ from "lodash";
import * as React from "react";
import { useHistory } from "react-router-dom";
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

export default function CoreClientContextProvider({ api, children }: Props) {
  const history = useHistory();

  const wrapped = {} as typeof Core;

  //   Wrap each API method in a check that redirects to the signin page if a 401 is returned.
  for (const method of Object.getOwnPropertyNames(api)) {
    if (typeof api[method] != 'function') {
      continue
    }
    wrapped[method] = (req, initReq) => {
      return api[method](req, initReq).catch((err) => {
        if (err.code === 401) {
          history.push(AuthRoutes.AUTH_PATH_SIGNIN);
        }
        throw err;
      });
    };
  }

  return (
    <CoreClientContext.Provider value={{ api: wrapped }}>
      {children}
    </CoreClientContext.Provider>
  );
}
