import { createMuiTheme, MuiThemeProvider } from "@material-ui/core";
import * as React from "react";
import {
  BrowserRouter as Router,
  Redirect,
  Route,
  Switch,
} from "react-router-dom";
import AuthenticatedRoute from "./components/AuthenticatedRoute";
import { AuthProvider } from "./contexts/AuthContext";
import Login from "./pages/Login";
import User from "./pages/User";

export default function App() {
  return (
    <div>
      <div>Weave GitOps UI</div>
      <MuiThemeProvider theme={createMuiTheme({})}>
        <AuthProvider>
          <Router>
            <Switch>
              <Route exact path="/login" component={Login} />
              <AuthenticatedRoute exact path="/user" component={User} />
              <AuthenticatedRoute
                exact
                path="/add_app"
                component={() => <h2>Add app</h2>}
              />
              <Redirect exact path="/callback" to="/user" />
              <Redirect exact path="/" to="/user" />
            </Switch>
          </Router>
        </AuthProvider>
      </MuiThemeProvider>
    </div>
  );
}
