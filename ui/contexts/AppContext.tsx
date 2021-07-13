import _ from "lodash";
import * as React from "react";
import { useHistory } from "react-router-dom";
import { Applications, User } from "../lib/api/applications/applications.pb";
import { getToken } from "../lib/storage";
import { PageRoute } from "../lib/types";

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

export type AppContextType = {
  applicationsClient: typeof Applications;
  user: null | User;
  loading: boolean;
  setUser: (user: User) => void;
};

export const AppContext = React.createContext<AppContextType>(
  null as AppContextType
);

export default function AppContextProvider({ applicationsClient, ...props }) {
  const history = useHistory();
  const [user, setUser] = React.useState(null);
  const [loading, setLoading] = React.useState(true);

  const appsClient = wrapClient<typeof Applications>(applicationsClient);

  React.useEffect(() => {
    // Fetch the user once at app startup
    appsClient
      .GetUser({})
      .then(({ user }) => {
        setUser(user);
      })
      .catch((err) => {
        console.error(err);
        history.push(PageRoute.Auth);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  const value: AppContextType = {
    applicationsClient: appsClient,
    user,
    loading,
    setUser,
  };

  return <AppContext.Provider {...props} value={value} />;
}
