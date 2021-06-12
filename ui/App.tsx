import { MuiThemeProvider } from "@material-ui/core";
import * as React from "react";
import {
  BrowserRouter as Router,
  Redirect,
  Route,
  Switch,
} from "react-router-dom";
import { ThemeProvider } from "styled-components";
import ErrorBoundary from "./components/ErrorBoundary";
import theme, { GlobalStyle, muiTheme } from "./lib/theme";
import { PageRoute } from "./lib/types";
import ApplicationDetail from "./pages/ApplicationDetail";
import Applications from "./pages/Applications";
import Error from "./pages/Error";

export default function App() {
  return (
    <MuiThemeProvider theme={muiTheme}>
      <ThemeProvider theme={theme}>
        <GlobalStyle />
        <Router>
          <ErrorBoundary>
            <Switch>
              <Route
                exact
                path={PageRoute.Applications}
                component={Applications}
              />
              <Route
                exact
                path={PageRoute.ApplicationDetail}
                component={ApplicationDetail}
              />
              <Redirect exact from="/" to={PageRoute.Applications} />
              <Route exact path="*" component={Error} />
            </Switch>
          </ErrorBoundary>
        </Router>
      </ThemeProvider>
    </MuiThemeProvider>
  );
}
