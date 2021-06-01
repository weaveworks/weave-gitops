import { Button, createMuiTheme, MuiThemeProvider } from "@material-ui/core";
import * as React from "react";
import { BrowserRouter as Router, Route, Switch } from "react-router-dom";
import { Redirect } from "react-router-dom/cjs/react-router-dom.min";
import AuthenticatedRoute from "./components/AuthenticatedRoute";
import { AuthProvider } from "./contexts/AuthContext";
import { DefaultGitOps } from "./lib/rpc/gitops";
import { wrappedFetch } from "./lib/util";
import User from "./pages/User";

const gitopsClient = new DefaultGitOps("/api/gitops", wrappedFetch);

export default function App() {
  const [ok, setOk] = React.useState(false);
  const [error, setError] = React.useState(null);

  const handleLogin = () => {
    gitopsClient.login({ state: "" }).then((res) => {
      window.location.href = res.redirectUrl;
    });
  };

  return (
    <div>
      <div>Weave GitOps UI</div>
      <MuiThemeProvider theme={createMuiTheme({})}>
        <AuthProvider>
          <Router>
            <Switch>
              <Route
                exact
                path="/login"
                component={() => (
                  <Button onClick={handleLogin} color="primary">
                    Login
                  </Button>
                )}
              />
              <AuthenticatedRoute exact path="/user" component={User} />
              <Redirect exact path="/callback" to="/user" />
              <Redirect exact path="/" to="/user" />
            </Switch>
          </Router>
        </AuthProvider>
      </MuiThemeProvider>
    </div>
  );
}
