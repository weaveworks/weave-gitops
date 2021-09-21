import _ from "lodash";
import * as React from "react";
import { Applications } from "../lib/api/applications/applications.pb";
import { GitProviderName } from "../lib/types";
import { getProviderToken, storeProviderToken } from "../lib/utils";

type AppState = {
  error: null | { fatal: boolean; message: string; detail?: string };
};

export type LinkResolver = (incoming: string) => string;

export function defaultLinkResolver(incoming: string): string {
  return incoming;
}

export type AppContextType = {
  applicationsClient: typeof Applications;
  doAsyncError: (message: string, detail: string) => void;
  appState: AppState;
  linkResolver: LinkResolver;
  getProviderToken: typeof getProviderToken;
  storeProviderToken: typeof storeProviderToken;
};

export const AppContext = React.createContext<AppContextType>(
  null as AppContextType
);

export interface Props {
  applicationsClient: typeof Applications;
  linkResolver?: LinkResolver;
  children?: any;
}

// Due to the way the grpc-gateway typescript client is generated,
// we need to wrap each individual call to ensure the auth header gets
// injected into the underlying `fetch` requests.
// This saves us from having to rememeber to pass it as an arg in every request.
function wrapClient<T>(client: any, tokenGetter: () => string): T {
  const wrapped = {};

  _.each(client, (func, name) => {
    wrapped[name] = (payload, options: RequestInit = {}) => {
      const withToken: RequestInit = {
        ...options,
        headers: new Headers({
          ...(options.headers || {}),
          Authorization: `token ${tokenGetter()}`,
        }),
      };

      return func(payload, withToken);
    };
  });

  return wrapped as T;
}

export default function AppContextProvider({
  applicationsClient,
  ...props
}: Props) {
  const [appState, setAppState] = React.useState({
    error: null,
  });

  React.useEffect(() => {
    // clear the error state on navigation
    setAppState({
      ...appState,
      error: null,
    });
  }, [window.location]);

  const doAsyncError = (message: string, detail: string) => {
    console.error(message);
    setAppState({
      ...appState,
      error: { message, detail },
    });
  };

  const value: AppContextType = {
    applicationsClient: wrapClient(applicationsClient, () =>
      getProviderToken(GitProviderName.GitHub)
    ),
    doAsyncError,
    appState,
    linkResolver: props.linkResolver || defaultLinkResolver,
    getProviderToken,
    storeProviderToken,
  };

  return <AppContext.Provider {...props} value={value} />;
}
