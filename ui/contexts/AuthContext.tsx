import qs from "query-string";
import * as React from "react";
import { Navigate, useLocation, useNavigate } from "react-router";
import { reloadBrowserSignIn } from "../lib/utils";
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
  const { userInfo } = React.useContext(Auth) as AuthContext;
  const location = useLocation();
  // Wait until userInfo is loaded before showing signin or app content
  if (!userInfo) {
    return Loader ? Loader : null;
  }
  // Signed in! Show app
  if (userInfo?.id) {
    return children;
  }
  // User appears not be logged in, off to signin
  return (
    <Navigate
      to={{
        pathname: AuthRoutes.AUTH_PATH_SIGNIN,
        search: qs.stringify({ redirect: location.pathname + location.search }),
      }}
    />
  );
};

export type AuthContext = {
  signIn: (data: any) => void;
  userInfo: {
    email: string;
    groups: string[];
    id: string;
  };
  error: { status: number; statusText: string };
  setError: any;
  loading: boolean;
  logOut: () => void;
};

export const Auth = React.createContext<AuthContext | null>({} as AuthContext);

export default function AuthContextProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const { request } = React.useContext(AppContext);

  const [userInfo, setUserInfo] = React.useState<{
    email: string;
    groups: string[];
    id: string;
  } | null>(null);
  const [loading, setLoading] = React.useState<boolean>(true);
  const [error, setError] = React.useState<Response | null>(null);
  const navigate = useNavigate();
  const location = useLocation();

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
      .then((data) =>
        setUserInfo({ email: data?.email, groups: data?.groups, id: data?.id }),
      )
      .catch((err) => console.log(err))
      .finally(() => setLoading(false));
  }, [request]);

  const signIn = React.useCallback(
    (data: any) => {
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
            navigate(qs.parse(location.search).redirect?.toString() || "/");
          });
        })
        .finally(() => setLoading(false));
    },
    [getUserInfo, location, navigate, request],
  );

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
        reloadBrowserSignIn();
      })
      .finally(() => setLoading(false));
  }, [request]);

  React.useEffect(() => {
    getUserInfo();
  }, [getUserInfo, location]);

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
