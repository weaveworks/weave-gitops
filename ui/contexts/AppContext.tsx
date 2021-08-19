import * as React from "react";
import { Applications } from "../lib/api/applications/applications.pb";

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
};

export const AppContext = React.createContext<AppContextType>(
  null as AppContextType
);

export interface Props {
  applicationsClient: typeof Applications;
  linkResolver?: LinkResolver;
  children?: any;
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
    applicationsClient,
    doAsyncError,
    appState,
    linkResolver: props.linkResolver || defaultLinkResolver,
  };

  return <AppContext.Provider {...props} value={value} />;
}
