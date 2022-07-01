import * as React from "react";
import { Redirect, useHistory } from "react-router-dom";
import { AppContext } from "./AppContext";

export enum AuthRoutes {
  USER_INFO = "/oauth2/userinfo",
  SIGN_IN = "/oauth2/sign_in",
  LOG_OUT = "/oauth2/logout",
  AUTH_PATH_SIGNIN = "/sign_in",
}

interface AuthCheckProps {
  children: any;
  Loader?: React.ElementType;
}

export const AuthCheck = ({ children, Loader }: AuthCheckProps) => {
  const { userInfo } = React.useContext(Auth);

  // Wait until userInfo is loaded before showing signin or app content
  if (!userInfo) {
    return Loader ? <Loader /> : null;
  }

  // Signed in! Show app
  if (userInfo?.email) {
    return children;
  }

  // User appears not be logged in, off to signin
  return <Redirect to={AuthRoutes.AUTH_PATH_SIGNIN} />;
};

export type AuthContext = {
  signIn: (data: any) => void;
  userInfo: {
    email: string;
    groups: string[];
  };
  error: { status: number; statusText: string };
  setError: any;
  loading: boolean;
  logOut: () => void;
};

export const Auth = React.createContext<AuthContext | null>({} as AuthContext);

export default function AuthContextProvider({ children }) {
  const { request } = React.useContext(AppContext);

  const [userInfo, setUserInfo] = React.useState<{
    email: string;
    groups: string[];
  }>(null);
  const [loading, setLoading] = React.useState<boolean>(true);
  const [error, setError] = React.useState(null);
  const history = useHistory();

  const signIn = React.useCallback((data) => {
    setLoading(true);
    request(AuthRoutes.SIGN_IN, {
      method: "POST",
      body: JSON.stringify(data),
    })
      .then((response) => {
        if (response.status !== 200) {
          setError(response);
          return;
        }
        getUserInfo().then(() => {
          setError(null);
          const prev = history.location.state;
          if (prev) history.push(prev.pathname + prev.search);
          else history.push("/");
        });
      })
      .finally(() => setLoading(false));
  }, []);

  const getUserInfo = React.useCallback(() => {
    setLoading(true);
    return request(AuthRoutes.USER_INFO)
      .then((response) => {
        if (response.status === 400 || response.status === 401) {
          setUserInfo(null);
          return;
        }
        return response.json();
      })
      .then((data) => setUserInfo({ email: data?.email, groups: [] }))
      .catch((err) => console.log(err))
      .finally(() => setLoading(false));
  }, []);

  const logOut = React.useCallback(() => {
    setLoading(true);
    request(AuthRoutes.LOG_OUT, {
      method: "POST",
    })
      .then((response) => {
        if (response.status !== 200) {
          setError(response);
          return;
        }
        window.location.pathname = AuthRoutes.AUTH_PATH_SIGNIN;
      })
      .finally(() => setLoading(false));
  }, []);

  React.useEffect(() => {
    getUserInfo();
    return history.listen(getUserInfo);
  }, [getUserInfo, history]);

  return (
    <>
      <Auth.Provider
        value={{
          signIn,
          userInfo,
          error,
          setError,
          loading,
          logOut,
        }}
      >
        {children}
      </Auth.Provider>
    </>
  );
}
