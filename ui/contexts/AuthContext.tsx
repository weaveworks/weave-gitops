import * as React from "react";
import { Redirect, Route, Switch } from "react-router-dom";
import SignIn from "../pages/SignIn";

const AUTH_PATH_SIGNIN = "/sign_in";
const AUTH_PATH_RESET_PASSWORD = "/reset_password";
const USER_INFO = "/user_info";

export const AuthSwitch: React.FC = () => {
  const handle404 = () => {
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
  submitAuthType: (selection: string) => void;
  userInfo: {
    email: string;
    groups: string[];
  };
};

export const Auth = React.createContext<AuthContext | null>(null);

export default function AuthContextProvider({ children }) {
  const [cookie, setCookie] = React.useState(null);
  const [userInfo, setUserInfo] = React.useState<{
    email: string;
    groups: string[];
  } | null>(null);

  React.useEffect(() => {
    // clear the error state on navigation

    getUserInfo();
  }, [window.location]);

  const submitAuthType = (selection: string) => {
    // POST user selection => user/pass OR OIDC
  };

  const getUserInfo = () => {
    // call API here
    // set state for user Info
    // if 401 => user not authenticated => leave null
  };

  return (
    <Auth.Provider value={{ submitAuthType, userInfo }}>
      {cookie && userInfo ? children : <AuthSwitch />}
    </Auth.Provider>
  );
}
