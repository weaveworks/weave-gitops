import _ from "lodash";
import * as React from "react";
import { Applications } from "../lib/api/applications/applications.pb";
import {
  clearCallbackState,
  getCallbackState,
  getProviderToken,
  storeCallbackState,
  storeProviderToken,
} from "../lib/storage";
import { notifySuccess } from "../lib/utils";

type AppState = {
  error: null | { fatal: boolean; message: string; detail?: string };
};

type AppSettings = {
  renderFooter: boolean;
};

export type LinkResolver = (incoming: string) => string;

export function defaultLinkResolver(incoming: string): string {
  return incoming;
}

export type AppContextType = {
  applicationsClient: typeof Applications;
  doAsyncError: (message: string, detail: string) => void;
  appState: AppState;
  settings: AppSettings;
  linkResolver: LinkResolver;
  getProviderToken: typeof getProviderToken;
  storeProviderToken: typeof storeProviderToken;
  getCallbackState: typeof getCallbackState;
  storeCallbackState: typeof storeCallbackState;
  clearCallbackState: typeof clearCallbackState;
  navigate: (url: string) => void;
  notifySuccess: typeof notifySuccess;
};

export const AppContext = React.createContext<AppContextType>(
  null as AppContextType
);

export interface AppProps {
  applicationsClient?: typeof Applications;
  linkResolver?: LinkResolver;
  children?: any;
  renderFooter?: boolean;
  notifySuccess?: typeof notifySuccess;
}

// Due to the way the grpc-gateway typescript client is generated,
// we need to wrap each individual call to ensure the auth header gets
// injected into the underlying `fetch` requests.
// This saves us from having to rememeber to pass it as an arg in every request.
function wrapClient<T>(client: any, tokenGetter: () => string): T {
  const wrapped = {};
  const gitProviderTokenHeader = "Git-Provider-Token";

  _.each(client, (func, name) => {
    wrapped[name] = (payload, options: RequestInit = {}) => {
      const withToken: RequestInit = {
        ...options,
        headers: new Headers({
          ...(options.headers || {}),
          [gitProviderTokenHeader]: `token ${tokenGetter()}`,
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
}: AppProps) {
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
    applicationsClient,
    doAsyncError,
    appState,
    linkResolver: props.linkResolver || defaultLinkResolver,
    getProviderToken,
    storeProviderToken,
    storeCallbackState,
    getCallbackState,
    clearCallbackState,
    notifySuccess: props.notifySuccess || notifySuccess,
    settings: {
      renderFooter: props.renderFooter,
    },
    navigate: (url) => {
      if (process.env.NODE_ENV === "test") {
        return;
      }
      window.location.href = url;
    },
  };

  return <AppContext.Provider {...props} value={value} />;
}
