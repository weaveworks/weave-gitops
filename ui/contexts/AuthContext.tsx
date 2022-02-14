import * as React from "react";
import { Redirect, Route, Switch } from "react-router-dom";
import cookie from "react-cookie";
import SignIn from "../pages/SignIn";

const AUTH_PATH_SIGNIN = "/sign_in";
const USER_INFO = "oauth2/userinfo";
const API_URL = process.env.REACT_API_URL as string;

export type RequestMethod =
  | "GET"
  | "HEAD"
  | "POST"
  | "PUT"
  | "DELETE"
  | "CONNECT"
  | "OPTIONS"
  | "TRACE";

export const processResponse = (res: Response) => {
  // 400s / 500s have res.ok = false
  if (!res.ok) {
    return res
      .clone()
      .json()
      .catch(() => res.text().then((message) => ({ message })))
      .then((data) => Promise.reject(data));
  }
  return res
    .clone()
    .json()
    .catch(() => res.text().then((message) => ({ success: true, message })));
};

export const request = (
  method: RequestMethod,
  query: RequestInfo,
  options: RequestInit = {}
) =>
  window
    .fetch(query, { ...options, method })
    .then((res) => processResponse(res));

export const AuthSwitch: React.FC = () => {
  const handle404 = () => {
    console.log(window.location);
    console.log("getting 404");
    return <Redirect to={AUTH_PATH_SIGNIN} />;
  };

  return (
    <Switch>
      <Route component={SignIn} exact={true} path={AUTH_PATH_SIGNIN} />
      {/* <Route
        component={PageResetPassword}
        exact={true}
        path={AUTH_PATH_RESET_PASSWORD}
      /> */}
      <Route render={handle404} />
    </Switch>
  );
};

export type AuthContext = {
  signIn: (username?: string, password?: string) => void;
  userInfo: {
    email: string;
    groups: string[];
  };
};

export const Auth = React.createContext<AuthContext | null>(null);

export default function AuthContextProvider({ children }) {
  const [userInfo, setUserInfo] = React.useState<{
    email: string;
    groups: string[];
  } | null>(null);

  const signIn = React.useCallback((username?: string, password?: string) => {
    fetch(
      `${API_URL}/oauth2/sign_in?return_url=${encodeURIComponent(
        "0.0.0.0:4567"
      )}`,
      {
        method: "POST",
        body: JSON.stringify({ username, password }),
      }
    )
      .then((res) => console.log(res))
      .catch((err) => console.log(err));
  }, []);

  const getUserInfo = React.useCallback(() => {
    fetch(`${API_URL}/${USER_INFO}`, {
      credentials: "include",
    })
      .then((res) => {
        console.log(res);
        setUserInfo(res.body as any);
      })
      .catch((err) => {
        console.log(err);
        if (err.code === "401") {
          // user is not authenticated
        }
      });
    // set state for user Info
    // if 401 => user not authenticated => leave null
  }, []);

  console.log(userInfo);

  React.useEffect(() => {
    getUserInfo();
  }, []);

  return (
    <Auth.Provider value={{ signIn, userInfo }}>
      {userInfo?.email ? children : <AuthSwitch />}
    </Auth.Provider>
  );
}
