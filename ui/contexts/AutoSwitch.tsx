import * as React from "react";
import { Redirect, Route, Switch } from "react-router-dom";
import SignIn from "../pages/SignIn";

const AUTH_PATH_SIGNIN = "/sign_in";

export const AuthSwitch: React.FC = () => {
  const handle404 = () => <Redirect exact={true} to={AUTH_PATH_SIGNIN} />;

  console.log("In the Auth Switch");

  return (
    <Switch>
      <Route component={SignIn} exact={true} path={AUTH_PATH_SIGNIN} />
      <Route render={handle404} />
    </Switch>
  );
};
