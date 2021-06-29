import * as React from "react";
import { Applications } from "../lib/api/applications/applications.pb";

type AppState = {
  error: null | { fatal: boolean; message: string; detail?: string };
};

export type AppContextType = {
  applicationsClient: typeof Applications;
  doAsyncError: (message: string, detail: string) => void;
  appState: AppState;
};

export const AppContext = React.createContext<AppContextType>(
  null as AppContextType
);

export default function AppContextProvider({ applicationsClient, ...props }) {
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
  };

  return <AppContext.Provider {...props} value={value} />;
}
