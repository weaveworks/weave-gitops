import _ from "lodash";
import * as React from "react";
import { Applications, User } from "../lib/api/applications/applications.pb";
import { getToken } from "../lib/storage";

type AppState = {
  error: null | { fatal: boolean; message: string; detail?: string };
};

export type AppContextType = {
  applicationsClient: typeof Applications;
  doAsyncError: (message: string, detail: string) => void;
  appState: AppState;
  user: User;
  loading: boolean;
};

// Due to the way the grpc-gateway typescript client is generated,
// we need to wrap each individual call to ensure the auth header gets
// injected into the underlying `fetch` requests.
// This saves us from having to rememeber to pass it as an arg in every request.
function wrapClient<T>(client: any): T {
  const wrapped = {};

  _.each(client, (func, name) => {
    wrapped[name] = (payload, options: RequestInit = {}) => {
      const token = getToken();

      const withToken: RequestInit = {
        ...options,
        headers: new Headers({
          ...(options.headers || {}),
          Authorization: `token ${token}`,
        }),
      };

      return func(payload, withToken);
    };
  });

  return wrapped as T;
}

export const AppContext = React.createContext<AppContextType>(
  null as AppContextType
);

export default function AppContextProvider({ applicationsClient, ...props }) {
  const [loading, setLoading] = React.useState(true);
  const [user, setUser] = React.useState<User>();
  const [appState, setAppState] = React.useState({
    error: null,
  });

  const appsClient = wrapClient<typeof Applications>(applicationsClient);
  React.useEffect(() => {
    // clear the error state on navigation
    setAppState({
      ...appState,
      error: null,
    });
  }, [window.location]);

  React.useEffect(() => {
    appsClient
      .GetUser({})
      .then((res) => {
        setUser(res.user);
      })
      .finally(() => setLoading(false));
  }, []);

  const doAsyncError = (message: string, detail: string) => {
    console.error(message);
    setAppState({
      ...appState,
      error: { message, detail },
    });
  };

  const value: AppContextType = {
    applicationsClient: appsClient,
    doAsyncError,
    appState,
    user,
    loading,
  };

  return <AppContext.Provider {...props} value={value} />;
}
