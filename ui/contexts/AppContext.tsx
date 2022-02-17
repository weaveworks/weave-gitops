import * as React from "react";
import { useHistory } from "react-router-dom";
import { Applications } from "../lib/api/applications/applications.pb";
import {
  clearCallbackState,
  getCallbackState,
  getProviderToken,
  storeCallbackState,
  storeProviderToken,
} from "../lib/storage";
import { PageRoute } from "../lib/types";
import { formatURL, notifySuccess } from "../lib/utils";

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
  clearAsyncError: () => void;
  appState: AppState;
  settings: AppSettings;
  linkResolver: LinkResolver;
  getProviderToken: typeof getProviderToken;
  storeProviderToken: typeof storeProviderToken;
  getCallbackState: typeof getCallbackState;
  storeCallbackState: typeof storeCallbackState;
  clearCallbackState: typeof clearCallbackState;
  navigate: {
    internal: (page: PageRoute, query?: any) => void;
    external: (url: string) => void;
  };
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

export default function AppContextProvider({
  applicationsClient,
  ...props
}: AppProps) {
  const history = useHistory();
  const [appState, setAppState] = React.useState({
    error: null,
  });

  const clearAsyncError = () => {
    setAppState({
      ...appState,
      error: null,
    });
  };

  React.useEffect(() => {
    // clear the error state on navigation
    clearAsyncError();
  }, [window.location]);

  console.log(window.location);

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
    clearAsyncError,
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
    navigate: {
      internal: (page: PageRoute, query?: any) => {
        const u = formatURL(page, query);

        history.push(u);
      },
      external: (url) => {
        if (process.env.NODE_ENV === "test") {
          return;
        }
        window.location.href = url;
      },
    },
  };

  return <AppContext.Provider {...props} value={value} />;
}
